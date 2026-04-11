package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	usersGRPC "booker/cmd/grpc/users"
	"booker/cmd/shared"
	"booker/config"
	userServices "booker/modules/users/application/services"
	userUsecases "booker/modules/users/application/usecases"
	userRepos "booker/modules/users/infrastructure/repositories"
	"booker/modules/users/infrastructure/token"
	"booker/pkg/gateway"
	"booker/pkg/interceptors"
	bookerOtel "booker/pkg/otel"
	pb "booker/proto/user/v1/gen"

	urfavecli "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// RunUsersSvc starts the users/auth service (gRPC + optional grpc-gateway REST).
func RunUsersSvc(c *urfavecli.Context) error {
	configPath := c.String("config")
	port := c.Int("port")
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

	// Create gRPC server with OTel + app interceptors
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingUnaryInterceptor(log),
			interceptors.UserHeaderInterceptor(),
		),
	)
	pb.RegisterUserServiceServer(grpcServer, usersGRPC.NewUserServer(userService))
	pb.RegisterAuthServiceServer(grpcServer, usersGRPC.NewAuthServer(
		registerUC, loginUC, refreshTokenUC, logoutUC, tokenService, userService,
	))

	// Health check + reflection
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user.v1.UserService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("user.v1.AuthService", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	// Listen gRPC
	if port == 0 {
		port = 50051
	}
	grpcAddr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	go func() {
		log.With("address", grpcAddr).Info("Users/Auth gRPC started")
		if err := grpcServer.Serve(lis); err != nil {
			log.With("error", err.Error()).Error("gRPC server failed")
		}
	}()

	// Start grpc-gateway REST API
	var httpServer *http.Server
	if httpPort > 0 {
		gwMux := gateway.NewGatewayMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, gwMux, grpcAddr, opts); err != nil {
			return fmt.Errorf("failed to register user gateway: %w", err)
		}
		if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, gwMux, grpcAddr, opts); err != nil {
			return fmt.Errorf("failed to register auth gateway: %w", err)
		}

		// ForwardAuth endpoint for Traefik
		topMux := http.NewServeMux()
		topMux.HandleFunc("/auth/verify", func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			claims, err := tokenService.ValidateAccessToken(r.Context(), authHeader[7:])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("X-User-Id", claims.UserID)
			w.Header().Set("X-Role", claims.Role)
			w.WriteHeader(http.StatusOK)
		})
		topMux.Handle("/", gwMux)

		httpAddr := fmt.Sprintf(":%d", httpPort)
		httpServer = gateway.NewHTTPServer(httpAddr, topMux, log)
		go func() {
			log.With("address", httpAddr).Info("Users/Auth REST API started (grpc-gateway)")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.With("error", err.Error()).Error("REST API failed")
			}
		}()
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down users service...")
	healthServer.SetServingStatus("user.v1.UserService", healthpb.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus("user.v1.AuthService", healthpb.HealthCheckResponse_NOT_SERVING)
	if httpServer != nil {
		_ = httpServer.Close()
	}
	grpcServer.GracefulStop()

	return nil
}
