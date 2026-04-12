package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	notifHTTP "booker/cmd/http/notification"
	"booker/cmd/shared"
	"booker/config"
	notifServices "booker/modules/notification/application/services"
	"booker/modules/notification/infrastructure/consumer"
	notifRepos "booker/modules/notification/infrastructure/repositories"
	"booker/modules/notification/infrastructure/ws"
	"booker/modules/users/infrastructure/token"
	"booker/pkg/httpserver"
	pkgnats "booker/pkg/nats"
	bookerOtel "booker/pkg/otel"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	urfavecli "github.com/urfave/cli/v2"
)

// RunNotificationSvc starts the notification service (Fiber REST + WebSocket + NATS consumers).
func RunNotificationSvc(c *urfavecli.Context) error {
	configPath := c.String("config")
	httpPort := c.Int("http-port")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := shared.InitLogger(cfg)
	ctx := context.Background()

	otelShutdown, err := bookerOtel.Setup(ctx, bookerOtel.Config{
		ServiceName: "notification-svc",
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

	// Wire notification module
	notifRepo := notifRepos.NewNotificationRepository(db)
	wsHub := ws.NewHub()
	notifService := notifServices.NewNotificationService(notifRepo, wsHub)

	// NATS JetStream consumers
	var natsConsumer *consumer.NATSConsumer
	nc, js, natsErr := shared.InitNATS(cfg.NATS.URL)
	if natsErr != nil {
		log.With("error", natsErr.Error()).Warn("NATS unavailable, real-time event consumption disabled")
	} else {
		defer nc.Close()
		if err := pkgnats.EnsureStreams(js); err != nil {
			log.With("error", err.Error()).Warn("failed to ensure NATS streams")
		}
		log.Info("NATS JetStream connected")

		eventHandler := consumer.NewEventHandler(notifService)
		natsConsumer = consumer.NewNATSConsumer(js, eventHandler, log)
		consumerCtx, consumerCancel := context.WithCancel(ctx)
		defer consumerCancel()
		natsConsumer.Start(consumerCtx)
	}

	// Token service for auth middleware
	tokenService := token.NewJWTTokenService(redisClient, cfg.JWT)

	// --- Fiber REST + WebSocket server ---
	app := fiber.New(fiber.Config{
		ErrorHandler: httpserver.ErrorHandler,
		AppName:      "booker-notification-svc",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CorsOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-Id",
	}))
	app.Use(httpserver.RequestIDMiddleware())
	app.Use(httpserver.TracingMiddleware())
	app.Use(httpserver.LoggingMiddleware())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	notifHTTP.RegisterRoutes(app, notifService, tokenService, wsHub)

	if httpPort == 0 {
		httpPort = 8086
	}
	httpserver.LogRoutes(app, "notification-svc")
	httpAddr := fmt.Sprintf(":%d", httpPort)
	errCh := make(chan error, 1)
	go func() {
		log.With("address", httpAddr).Info("Notification service started (REST + WebSocket)")
		if err := app.Listen(httpAddr); err != nil {
			errCh <- fmt.Errorf("Fiber server failed: %w", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		log.Info("shutting down notification service...")
	case err := <-errCh:
		log.With("error", err.Error()).Error("server error, shutting down...")
	}

	if natsConsumer != nil {
		natsConsumer.Stop()
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("http shutdown error", "error", err)
	}

	return nil
}
