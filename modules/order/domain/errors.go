package domain

import (
	"net/http"

	apperrors "booker/pkg/errors"
)

var (
	ErrOrderNotFound = &apperrors.BaseAppError{
		Code: "ORDER_NOT_FOUND", Msg: "Order not found", HttpStatus: http.StatusNotFound,
	}
	ErrOrderNotCancellable = &apperrors.BaseAppError{
		Code: "ORDER_NOT_CANCELLABLE", Msg: "Order cannot be cancelled", HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidSide = &apperrors.BaseAppError{
		Code: "INVALID_SIDE", Msg: "Side must be buy or sell", HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidOrderType = &apperrors.BaseAppError{
		Code: "INVALID_ORDER_TYPE", Msg: "Only limit orders supported", HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidPrice = &apperrors.BaseAppError{
		Code: "INVALID_PRICE", Msg: "Price must be positive", HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidQuantity = &apperrors.BaseAppError{
		Code: "INVALID_QUANTITY", Msg: "Quantity must be positive", HttpStatus: http.StatusBadRequest,
	}
	ErrBelowMinQty = &apperrors.BaseAppError{
		Code: "BELOW_MIN_QTY", Msg: "Quantity below minimum", HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidTickSize = &apperrors.BaseAppError{
		Code: "INVALID_TICK_SIZE", Msg: "Price must be a multiple of tick size", HttpStatus: http.StatusBadRequest,
	}
	ErrPairNotFound = &apperrors.BaseAppError{
		Code: "PAIR_NOT_FOUND", Msg: "Trading pair not found", HttpStatus: http.StatusNotFound,
	}
	ErrPairNotActive = &apperrors.BaseAppError{
		Code: "PAIR_NOT_ACTIVE", Msg: "Trading pair is not active", HttpStatus: http.StatusBadRequest,
	}
	ErrInsufficientBalance = &apperrors.BaseAppError{
		Code: "INSUFFICIENT_BALANCE", Msg: "Insufficient balance to place order", HttpStatus: http.StatusBadRequest,
	}
	ErrWalletServiceUnavailable = &apperrors.BaseAppError{
		Code: "WALLET_SERVICE_UNAVAILABLE", Msg: "Wallet service temporarily unavailable", HttpStatus: http.StatusServiceUnavailable,
	}
	ErrInvalidFillQty = &apperrors.BaseAppError{
		Code: "INVALID_FILL_QTY", Msg: "Fill quantity exceeds order quantity", HttpStatus: http.StatusBadRequest,
	}
	ErrOrderNotFillable = &apperrors.BaseAppError{
		Code: "ORDER_NOT_FILLABLE", Msg: "Order cannot be filled in current status", HttpStatus: http.StatusBadRequest,
	}
)
