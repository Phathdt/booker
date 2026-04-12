package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderGRPC "booker/cmd/grpc/order"
	orderHTTP "booker/cmd/http/order"
	"booker/cmd/shared"
	"booker/config"
	orderServices "booker/modules/order/application/services"
	orderInfra "booker/modules/order/infrastructure"
	orderRepos "booker/modules/order/infrastructure/repositories"
	"booker/modules/users/infrastructure/token"
	"booker/pkg/httpserver"
	"booker/pkg/interceptors"
	bookerOtel "booker/pkg/otel"
	pb "booker/proto/order/v1/gen"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	urfavecli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	_ "booker/docs"
)

// RunOrderSvc starts the order service (Fiber REST + gRPC).
func RunOrderSvc(c *urfavecli.Context) error {
	configPath := c.String("config")
	grpcPort := c.Int("port")
	httpPort := c.Int("http-port")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := shared.InitLogger(cfg)
	ctx := context.Background()

	otelShutdown, err := bookerOtel.Setup(ctx, bookerOtel.Config{
		ServiceName: "order-svc",
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

	redisClient, err := shared.InitRedis(ctx, cfg.Redis.URI)
	if err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}
	defer redisClient.Close()

	// Wallet gRPC client connection
	walletConn, err := grpc.NewClient(cfg.WalletService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to wallet-svc: %w", err)
	}
	defer walletConn.Close()

	// Wire order module
	orderRepo := orderRepos.NewOrderRepository(db)
	walletClient := orderInfra.NewWalletClient(walletConn)
	orderService := orderServices.NewOrderService(orderRepo, walletClient)

	// Token service for auth middleware
	tokenService := token.NewJWTTokenService(redisClient, cfg.JWT)

	// --- gRPC server (inter-service) ---
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingUnaryInterceptor(log),
		),
	)
	pb.RegisterOrderServiceServer(grpcServer, orderGRPC.NewOrderServer(orderService))

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("order.v1.OrderService", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	if grpcPort == 0 {
		grpcPort = 50053
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	errCh := make(chan error, 2)
	go func() {
		log.With("address", grpcAddr).Info("Order gRPC started (inter-service)")
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// --- Fiber REST server (external) ---
	app := fiber.New(fiber.Config{
		ErrorHandler: httpserver.ErrorHandler,
		AppName:      "booker-order-svc",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-Id",
	}))
	app.Use(httpserver.RequestIDMiddleware())
	app.Use(httpserver.TracingMiddleware())
	app.Use(httpserver.LoggingMiddleware())
	app.Get("/swagger/*", swagger.HandlerDefault)

	orderHTTP.RegisterRoutes(app, orderService, tokenService)

	httpAddr := fmt.Sprintf(":%d", httpPort)
	go func() {
		log.With("address", httpAddr).Info("Order REST API started (Fiber)")
		if err := app.Listen(httpAddr); err != nil {
			errCh <- fmt.Errorf("Fiber server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		log.Info("shutting down order service...")
	case err := <-errCh:
		log.With("error", err.Error()).Error("server error, shutting down...")
	}

	healthServer.SetServingStatus("order.v1.OrderService", healthpb.HealthCheckResponse_NOT_SERVING)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(shutdownCtx)
	grpcServer.GracefulStop()
	walletConn.Close()

	return nil
}
