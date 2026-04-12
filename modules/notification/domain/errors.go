package domain

import (
	"net/http"

	apperrors "booker/pkg/errors"
)

var ErrNotificationNotFound = &apperrors.BaseAppError{
	Code: "NOTIFICATION_NOT_FOUND", Msg: "Notification not found", HttpStatus: http.StatusNotFound,
}
