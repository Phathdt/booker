package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Details string `json:"details,omitempty"`
}

// CustomErrorHandler maps gRPC errors to standard error response format.
func CustomErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	st, ok := status.FromError(err)
	if !ok {
		st = status.New(codes.Internal, err.Error())
	}

	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	resp := errorResponse{
		Code:    st.Code().String(),
		Message: st.Message(),
		Status:  httpStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(resp)
}
