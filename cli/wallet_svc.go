package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	walletGRPC "booker/cmd/grpc/wallet"
	walletHTTP "booker/cmd/http/wallet"
	"booker/cmd/shared"
	"booker/config"
	"booker/modules/users/infrastructure/token"
	walletServices "booker/modules/wallet/application/services"
	walletRepos "booker/modules/wallet/infrastructure/repositories"
	"booker/pkg/httpserver"
	"booker/pkg/interceptors"
	bookerOtel "booker/pkg/otel"
	pb "booker/proto/wallet/v1/gen"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	urfavecli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	_ "booker/docs"
)

// RunWalletSvc starts the wallet service (Fiber REST + gRPC).
func RunWalletSvc(c *urfavecli.Context) error {
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
		ServiceName: "wallet-svc",
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

	// Wire wallet module
	walletRepo := walletRepos.NewWalletRepository(db)
	walletService := walletServices.NewWalletService(walletRepo)

	// Token service for auth middleware
	tokenService := token.NewJWTTokenService(redisClient, cfg.JWT)

	// --- gRPC server (inter-service) ---
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingUnaryInterceptor(log),
		),
	)
	pb.RegisterWalletServiceServer(grpcServer, walletGRPC.NewWalletServer(walletService))

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("wallet.v1.WalletService", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	if grpcPort == 0 {
		grpcPort = 50052
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	errCh := make(chan error, 2)
	go func() {
		log.With("address", grpcAddr).Info("Wallet gRPC started (inter-service)")
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// --- Fiber REST server (external) ---
	app := fiber.New(fiber.Config{
		ErrorHandler: httpserver.ErrorHandler,
		AppName:      "booker-wallet-svc",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-Id",
	}))
	app.Use(httpserver.RequestIDMiddleware())
	app.Use(httpserver.TracingMiddleware())
	app.Get("/swagger/*", swagger.HandlerDefault)

	walletHTTP.RegisterRoutes(app, walletService, tokenService)

	httpAddr := fmt.Sprintf(":%d", httpPort)
	go func() {
		log.With("address", httpAddr).Info("Wallet REST API started (Fiber)")
		if err := app.Listen(httpAddr); err != nil {
			errCh <- fmt.Errorf("Fiber server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		log.Info("shutting down wallet service...")
	case err := <-errCh:
		log.With("error", err.Error()).Error("server error, shutting down...")
	}

	healthServer.SetServingStatus("wallet.v1.WalletService", healthpb.HealthCheckResponse_NOT_SERVING)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(shutdownCtx)
	grpcServer.GracefulStop()

	return nil
}
