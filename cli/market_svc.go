package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	marketHTTP "booker/cmd/http/market"
	"booker/cmd/shared"
	"booker/config"
	"booker/modules/market/consumer"
	"booker/modules/market/ticker"
	"booker/modules/market/trades"
	"booker/modules/market/ws"
	matchingRepos "booker/modules/matching/infrastructure/repositories"
	"booker/pkg/httpserver"
	pkgnats "booker/pkg/nats"
	bookerOtel "booker/pkg/otel"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	urfavecli "github.com/urfave/cli/v2"

	_ "booker/docs"
)

// RunMarketSvc starts the market data service (REST + WebSocket).
func RunMarketSvc(c *urfavecli.Context) error {
	configPath := c.String("config")
	httpPort := c.Int("http-port")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := shared.InitLogger(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	otelShutdown, err := bookerOtel.Setup(ctx, bookerOtel.Config{
		ServiceName: "market-svc",
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

	// Load active trading pairs
	queries := matchingRepos.NewQueries(db)
	tradingPairs, err := queries.ListActiveTradingPairs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load trading pairs: %w", err)
	}

	// Create ticker aggregators + recent trades per pair
	tickers := make(map[string]*ticker.Aggregator)
	recentTrades := make(map[string]*trades.RecentTrades)
	var pairInfos []marketHTTP.PairInfo

	for _, pair := range tradingPairs {
		tickers[pair.ID] = ticker.NewAggregator(pair.ID)
		recentTrades[pair.ID] = trades.NewRecentTrades()
		pairInfos = append(pairInfos, marketHTTP.PairInfo{
			ID:         pair.ID,
			BaseAsset:  pair.BaseAsset,
			QuoteAsset: pair.QuoteAsset,
			MinQty:     pair.MinQty.String(),
			TickSize:   pair.TickSize.String(),
		})
	}

	log.With("pairs", len(tradingPairs)).Info("loaded trading pairs")

	// WebSocket Hub
	hub := ws.NewHub()
	go hub.Run(ctx)

	// NATS consumer
	nc, js, err := shared.InitNATS(cfg.NATS.URL)
	if err != nil {
		log.With("error", err.Error()).Warn("failed to init NATS, real-time updates disabled")
	} else {
		defer nc.Close()
		if err := pkgnats.EnsureStreams(js); err != nil {
			log.With("error", err.Error()).Warn("failed to ensure NATS streams")
		}

		tradeConsumer := consumer.NewTradeConsumer(tickers, recentTrades, hub)
		if err := tradeConsumer.Start(ctx, js); err != nil {
			log.With("error", err.Error()).Warn("failed to start trade consumer")
		} else {
			log.Info("NATS trade consumer started")
		}
	}

	// --- Fiber REST + WS server ---
	app := fiber.New(fiber.Config{
		ErrorHandler: httpserver.ErrorHandler,
		AppName:      "booker-market-svc",
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

	marketHTTP.RegisterRoutes(app, tickers, recentTrades, pairInfos, hub)

	httpAddr := fmt.Sprintf(":%d", httpPort)
	errCh := make(chan error, 1)
	go func() {
		log.With("address", httpAddr).Info("Market REST + WS API started (Fiber)")
		if err := app.Listen(httpAddr); err != nil {
			errCh <- fmt.Errorf("Fiber server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		log.Info("shutting down market service...")
	case err := <-errCh:
		log.With("error", err.Error()).Error("server error, shutting down...")
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = app.ShutdownWithContext(shutdownCtx)

	return nil
}
