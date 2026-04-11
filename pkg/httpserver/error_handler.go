package httpserver

import (
	"errors"

	apperrors "booker/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is Fiber's custom error handler that maps AppError to JSON.
func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	errResp := map[string]any{
		"code":    "INTERNAL_ERROR",
		"message": "Internal server error",
	}

	// Check for Fiber error
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		errResp["code"] = "HTTP_ERROR"
		errResp["message"] = fiberErr.Message
	}

	// Check for AppError
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		status = appErr.StatusCode()
		errResp["code"] = appErr.ErrorCode()
		errResp["message"] = appErr.Message()
		if appErr.Details() != "" {
			errResp["details"] = appErr.Details()
		}
	}

	return ErrorResponse(c, status, errResp)
}
