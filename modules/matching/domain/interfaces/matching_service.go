package interfaces

import (
	"context"

	"booker/modules/matching/engine"
)

type MatchingService interface {
	SubmitOrder(ctx context.Context, order *engine.BookOrder) ([]*engine.Trade, error)
	CancelOrder(ctx context.Context, pairID, orderID string) error
	GetOrderBook(ctx context.Context, pairID string) (*engine.OrderBookSnapshot, error)
}
