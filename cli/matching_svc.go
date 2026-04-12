package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	matchingGRPC "booker/cmd/grpc/matching"
	"booker/cmd/shared"
	"booker/config"
	matchingServices "booker/modules/matching/application/services"
	"booker/modules/matching/engine"
	matchingInfra "booker/modules/matching/infrastructure"
	matchingRepos "booker/modules/matching/infrastructure/repositories"
	"booker/pkg/interceptors"
	pkgnats "booker/pkg/nats"
	bookerOtel "booker/pkg/otel"
	pb "booker/proto/matching/v1/gen"

	urfavecli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// RunMatchingSvc starts the matching engine service (gRPC only).
func RunMatchingSvc(c *urfavecli.Context) error {
	configPath := c.String("config")
	grpcPort := c.Int("port")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := shared.InitLogger(cfg)
	ctx := context.Background()

	otelShutdown, err := bookerOtel.Setup(ctx, bookerOtel.Config{
		ServiceName: "matching-svc",
		Endpoint:    cfg.OTel.Endpoint,
		Insecure:    cfg.OTel.Insecure,
	})
	if err != nil {
		log.With("error", err.Error()).Warn("failed to init otel")
	} else {
		defer otelShutdown(ctx)
		log.Info("OpenTelemetry tracing initialized")
	}

	db, err := shared.InitDatabase(ctx, cfg.Database.URI, log)
	if err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer db.Close()

	// NATS JetStream
	nc, js, err := shared.InitNATS(cfg.NATS.URL)
	if err != nil {
		log.With("error", err.Error()).Warn("failed to init NATS, trade events disabled")
	} else {
		defer nc.Close()
		if err := pkgnats.EnsureStreams(js); err != nil {
			log.With("error", err.Error()).Warn("failed to ensure NATS streams")
		}
		log.Info("NATS JetStream connected")
	}

	// Wallet gRPC client
	walletConn, err := grpc.NewClient(cfg.WalletService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to wallet-svc: %w", err)
	}
	defer walletConn.Close()

	// Order gRPC client
	orderConn, err := grpc.NewClient(cfg.OrderService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to order-svc: %w", err)
	}
	defer orderConn.Close()

	// Wire matching module
	tradeRepo := matchingRepos.NewTradeRepository(db)
	walletClient := matchingInfra.NewWalletClient(walletConn)
	orderClient := matchingInfra.NewOrderClient(orderConn)

	var tradePublisher pkgnats.TradePublisher
	if js != nil {
		tradePublisher = pkgnats.NewTradePublisher(js)
	}

	// Load active trading pairs
	queries := matchingRepos.NewQueries(db)
	tradingPairs, err := queries.ListActiveTradingPairs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load trading pairs: %w", err)
	}

	// Create engines per pair + preload open orders (crash recovery)
	engines := make(map[string]*engine.Engine)
	pairs := make(map[string]*matchingServices.PairInfo)

	for _, pair := range tradingPairs {
		eng := engine.NewEngine(pair.ID, 1024)
		eng.Start(ctx)
		engines[pair.ID] = eng

		pairs[pair.ID] = &matchingServices.PairInfo{
			ID:         pair.ID,
			BaseAsset:  pair.BaseAsset,
			QuoteAsset: pair.QuoteAsset,
		}

		// Crash recovery: load open orders
		openOrders, err := queries.ListOpenOrdersByPair(ctx, pair.ID)
		if err != nil {
			log.With("pair", pair.ID, "error", err.Error()).Warn("failed to load open orders")
			continue
		}

		if len(openOrders) > 0 {
			bookOrders := make([]*engine.BookOrder, len(openOrders))
			for i, o := range openOrders {
				bookOrders[i] = &engine.BookOrder{
					ID:        o.ID,
					UserID:    o.UserID,
					PairID:    o.PairID,
					Side:      engine.Side(o.Side),
					Price:     o.Price,
					Quantity:  o.Quantity,
					Remaining: o.Quantity.Sub(o.FilledQty),
					CreatedAt: o.CreatedAt,
				}
			}
			eng.Preload(bookOrders)
			log.With("pair", pair.ID, "count", len(openOrders)).Info("preloaded open orders")
		}
	}

	matchingService := matchingServices.NewMatchingService(
		engines, tradeRepo, orderClient, walletClient, tradePublisher, pairs,
	)

	// --- gRPC server ---
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingUnaryInterceptor(log),
		),
	)
	pb.RegisterMatchingServiceServer(grpcServer, matchingGRPC.NewMatchingServer(matchingService))

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("matching.v1.MatchingService", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	if grpcPort == 0 {
		grpcPort = 50054
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.With("address", grpcAddr, "pairs", len(engines)).Info("Matching engine started (gRPC)")
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		log.Info("shutting down matching service...")
	case err := <-errCh:
		log.With("error", err.Error()).Error("server error, shutting down...")
	}

	healthServer.SetServingStatus("matching.v1.MatchingService", healthpb.HealthCheckResponse_NOT_SERVING)
	for _, eng := range engines {
		eng.Stop()
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info("gRPC server stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Warn("gRPC graceful stop timed out, forcing stop")
		grpcServer.Stop()
	}

	return nil
}
