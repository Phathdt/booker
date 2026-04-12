package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	usersGRPC "booker/cmd/grpc/users"
	usersHTTP "booker/cmd/http/users"
	"booker/cmd/shared"
	"booker/config"
	userServices "booker/modules/users/application/services"
	userUsecases "booker/modules/users/application/usecases"
	userRepos "booker/modules/users/infrastructure/repositories"
	"booker/modules/users/infrastructure/token"
	"booker/pkg/httpserver"
	"booker/pkg/interceptors"
	bookerOtel "booker/pkg/otel"
	pb "booker/proto/user/v1/gen"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	urfavecli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// RunUsersSvc starts the users/auth service (Fiber REST + gRPC).
func RunUsersSvc(c *urfavecli.Context) error {
	configPath := c.String("config")
	grpcPort := c.Int("port")
	httpPort := c.Int("http-port")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := shared.InitLogger(cfg)

	ctx := context.Background()

	// Init OpenTelemetry
	otelShutdown, err := bookerOtel.Setup(ctx, bookerOtel.Config{
		ServiceName: "users-svc",
		Endpoint:    cfg.OTel.Endpoint,
		Insecure:    cfg.OTel.Insecure,
	})
	if err != nil {
		log.With("error", err.Error()).Warn("failed to init otel, continuing without tracing")
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

	// Wire users module
	userRepo := userRepos.NewUserRepository(db)
	userService := userServices.NewUserService(userRepo)
	tokenService := token.NewJWTTokenService(redisClient, cfg.JWT)

	// Wire usecases
	registerUC := userUsecases.NewRegisterUseCase(userService, tokenService)
	loginUC := userUsecases.NewLoginUseCase(userService, tokenService)
	refreshTokenUC := userUsecases.NewRefreshTokenUseCase(userService, tokenService)
	logoutUC := userUsecases.NewLogoutUseCase(tokenService)

	// --- gRPC server (inter-service only) ---
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingUnaryInterceptor(log),
			interceptors.UserHeaderInterceptor(),
		),
	)
	pb.RegisterUserServiceServer(grpcServer, usersGRPC.NewUserServer(userService, tokenService))

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user.v1.UserService", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	if grpcPort == 0 {
		grpcPort = 50051
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	go func() {
		log.With("address", grpcAddr).Info("Users gRPC started (inter-service)")
		if err := grpcServer.Serve(lis); err != nil {
			log.With("error", err.Error()).Error("gRPC server failed")
		}
	}()

	// --- Fiber REST server (external) ---
	app := fiber.New(fiber.Config{
		ErrorHandler: httpserver.ErrorHandler,
		AppName:      "booker-users-svc",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CorsOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-Id",
		AllowCredentials: true,
	}))
	app.Use(httpserver.RequestIDMiddleware())
	app.Use(httpserver.TracingMiddleware())
	app.Use(httpserver.LoggingMiddleware())

	// ForwardAuth endpoint for Traefik
	app.Get("/auth/verify", func(fc *fiber.Ctx) error {
		authHeader := fc.Get("Authorization")
		if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			return fc.SendStatus(fiber.StatusUnauthorized)
		}
		claims, err := tokenService.ValidateAccessToken(fc.UserContext(), authHeader[7:])
		if err != nil {
			return fc.SendStatus(fiber.StatusUnauthorized)
		}
		fc.Set("X-User-Id", claims.UserID)
		fc.Set("X-Role", claims.Role)
		return fc.SendStatus(fiber.StatusOK)
	})

	// Register REST routes
	usersHTTP.RegisterRoutes(app, cfg, userService, tokenService, registerUC, loginUC, refreshTokenUC, logoutUC)

	httpserver.LogRoutes(app, "users-svc")
	httpAddr := fmt.Sprintf(":%d", httpPort)
	go func() {
		log.With("address", httpAddr).Info("Users REST API started (Fiber)")
		if err := app.Listen(httpAddr); err != nil {
			log.With("error", err.Error()).Error("Fiber server failed")
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down users service...")
	healthServer.SetServingStatus("user.v1.UserService", healthpb.HealthCheckResponse_NOT_SERVING)
	if err := app.Shutdown(); err != nil {
		log.Error("http shutdown error", "error", err)
	}
	grpcServer.GracefulStop()

	return nil
}
