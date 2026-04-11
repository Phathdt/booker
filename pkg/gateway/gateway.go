package gateway

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

// NewGatewayMux creates a grpc-gateway ServeMux with standard options:
// - snake_case JSON field names
// - {"data": ...} response wrapper
// - Custom error handler
// - Forwards x-user-id, x-role headers as gRPC metadata
func NewGatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithErrorHandler(CustomErrorHandler),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)
}

func headerMatcher(key string) (string, bool) {
	switch key {
	case "X-User-Id", "x-user-id":
		return "x-user-id", true
	case "X-Role", "x-role":
		return "x-role", true
	case "X-Request-Id", "x-request-id":
		return "x-request-id", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
