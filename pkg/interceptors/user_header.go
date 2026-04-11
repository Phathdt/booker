package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserHeaderInterceptor reads x-user-id and x-role from gRPC metadata
// (forwarded by grpc-gateway from HTTP headers set by Traefik ForwardAuth)
// and injects them into the Go context.
func UserHeaderInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if vals := md.Get("x-user-id"); len(vals) > 0 && vals[0] != "" {
				ctx = withUserID(ctx, vals[0])
			}
			role := "user"
			if vals := md.Get("x-role"); len(vals) > 0 && vals[0] != "" {
				role = vals[0]
			}
			ctx = withUserRole(ctx, role)
		}
		return handler(ctx, req)
	}
}
