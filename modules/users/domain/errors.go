package domain

import (
	"net/http"

	apperrors "booker/pkg/errors"
)

var (
	ErrEmailAlreadyExists = &apperrors.BaseAppError{
		Code: "EMAIL_EXISTS", Msg: "Email already exists", HttpStatus: http.StatusConflict,
	}
	ErrUserNotFound = &apperrors.BaseAppError{
		Code: "USER_NOT_FOUND", Msg: "User not found", HttpStatus: http.StatusNotFound,
	}
	ErrInvalidCredentials = &apperrors.BaseAppError{
		Code: "INVALID_CREDENTIALS", Msg: "Invalid email or password", HttpStatus: http.StatusUnauthorized,
	}
	ErrUserInactive = &apperrors.BaseAppError{
		Code: "USER_INACTIVE", Msg: "User account is not active", HttpStatus: http.StatusForbidden,
	}
	ErrInvalidToken = &apperrors.BaseAppError{
		Code: "INVALID_TOKEN", Msg: "Invalid or expired token", HttpStatus: http.StatusUnauthorized,
	}
)
