package domain

import (
	"net/http"

	apperrors "booker/pkg/errors"
)

var (
	ErrOrderNotInBook = &apperrors.BaseAppError{
		Code: "ORDER_NOT_IN_BOOK", Msg: "Order not found in book", HttpStatus: http.StatusNotFound,
	}
	ErrPairEngineNotFound = &apperrors.BaseAppError{
		Code: "PAIR_ENGINE_NOT_FOUND", Msg: "No matching engine for this pair", HttpStatus: http.StatusNotFound,
	}
	ErrSettlementFailed = &apperrors.BaseAppError{
		Code: "SETTLEMENT_FAILED", Msg: "Trade settlement failed", HttpStatus: http.StatusServiceUnavailable,
	}
	ErrOrderUpdateFailed = &apperrors.BaseAppError{
		Code: "ORDER_UPDATE_FAILED", Msg: "Failed to update order fill", HttpStatus: http.StatusServiceUnavailable,
	}
	ErrTradeNotFound = &apperrors.BaseAppError{
		Code: "TRADE_NOT_FOUND", Msg: "Trade not found", HttpStatus: http.StatusNotFound,
	}
)
