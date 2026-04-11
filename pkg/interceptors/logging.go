package interceptors

import (
	"context"
	"time"

	"booker/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor logs gRPC method, duration, and status code.
func LoggingUnaryInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		st, _ := status.FromError(err)
		entry := log.With(
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"code", st.Code().String(),
		)

		if err != nil {
			entry.With("error", err.Error()).Warn("gRPC call failed")
		} else {
			entry.Info("gRPC call")
		}

		return resp, err
	}
}
