package users

import (
	"errors"

	apperrors "booker/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toGRPCError maps domain AppError to gRPC status error.
func toGRPCError(err error) error {
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		code := httpStatusToGRPCCode(appErr.StatusCode())
		return status.Error(code, appErr.Message())
	}
	return status.Error(codes.Internal, "internal error")
}

func httpStatusToGRPCCode(httpStatus int) codes.Code {
	switch httpStatus {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}
