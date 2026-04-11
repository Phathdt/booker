package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// AppError represents application-level errors with HTTP context
type AppError interface {
	error
	StatusCode() int
	ErrorCode() string
	Message() string
	Details() string
}

// BaseAppError implements AppError interface
type BaseAppError struct {
	Code       string `json:"code"`
	Msg        string `json:"message"`
	Detail     string `json:"details,omitempty"`
	HttpStatus int    `json:"-"`
}

func (e *BaseAppError) Error() string {
	if e.Detail != "" {
		return e.Msg + ": " + e.Detail
	}
	return e.Msg
}

func (e *BaseAppError) StatusCode() int   { return e.HttpStatus }
func (e *BaseAppError) ErrorCode() string { return e.Code }
func (e *BaseAppError) Message() string   { return e.Msg }
func (e *BaseAppError) Details() string   { return e.Detail }

// Wrap returns a new error that preserves the root cause and captures a stack trace.
func (e *BaseAppError) Wrap(err error) AppError {
	return &WrappedAppError{
		BaseAppError: BaseAppError{
			Code:       e.Code,
			Msg:        e.Msg,
			Detail:     err.Error(),
			HttpStatus: e.HttpStatus,
		},
		rootCause: err,
		stack:     callers(2),
	}
}

// WrappedAppError extends BaseAppError with root cause and stack trace
type WrappedAppError struct {
	BaseAppError
	rootCause error
	stack     string
}

func (e *WrappedAppError) Error() string {
	return fmt.Sprintf("%s: %s\n%s", e.Msg, e.rootCause.Error(), e.stack)
}

func (e *WrappedAppError) Unwrap() error {
	return e.rootCause
}

func callers(skip int) string {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(skip+1, pcs)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}
		fmt.Fprintf(&b, "  at %s (%s:%d)\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return b.String()
}

// Common reusable errors
var (
	ErrMissingUserID = &BaseAppError{
		Code: "MISSING_USER_ID", Msg: "x-user-id header is required", HttpStatus: http.StatusBadRequest,
	}
)

// Common error constructors
func NewBadRequestError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusBadRequest}
}

func NewUnauthorizedError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusUnauthorized}
}

func NewNotFoundError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusNotFound}
}

func NewConflictError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusConflict}
}

func NewInternalError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusInternalServerError}
}
