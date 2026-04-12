package interfaces

import (
	"context"

	"booker/modules/order/domain/entities"

	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	GetTradingPair(ctx context.Context, pairID string) (*TradingPair, error)
	Create(ctx context.Context, order *entities.Order) (*entities.Order, error)
	GetByID(ctx context.Context, id string) (*entities.Order, error)
	GetByIDAndUser(ctx context.Context, id, userID string) (*entities.Order, error)
	List(ctx context.Context, userID string, pairID, status *string, limit, offset int32) ([]*entities.Order, error)
	Cancel(ctx context.Context, id, userID string) (*entities.Order, error)
	UpdateFilledQty(ctx context.Context, id string, filledQty decimal.Decimal, status string) (*entities.Order, error)
}

type TradingPair struct {
	ID         string
	BaseAsset  string
	QuoteAsset string
	Status     string
	MinQty     decimal.Decimal
	TickSize   decimal.Decimal
}
