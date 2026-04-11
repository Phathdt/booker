package domain

import (
	"net/http"

	apperrors "booker/pkg/errors"
)

var (
	ErrWalletNotFound = &apperrors.BaseAppError{
		Code: "WALLET_NOT_FOUND", Msg: "Wallet not found", HttpStatus: http.StatusNotFound,
	}
	ErrInsufficientBalance = &apperrors.BaseAppError{
		Code: "INSUFFICIENT_BALANCE", Msg: "Insufficient available balance", HttpStatus: http.StatusBadRequest,
	}
	ErrInsufficientLocked = &apperrors.BaseAppError{
		Code: "INSUFFICIENT_LOCKED", Msg: "Insufficient locked balance", HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidAmount = &apperrors.BaseAppError{
		Code: "INVALID_AMOUNT", Msg: "Amount must be positive", HttpStatus: http.StatusBadRequest,
	}
)
