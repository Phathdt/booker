package interfaces

import (
	"context"

	"booker/modules/matching/domain/entities"
)

type TradeRepository interface {
	Create(ctx context.Context, trade *entities.Trade) (*entities.Trade, error)
	GetByID(ctx context.Context, id string) (*entities.Trade, error)
	ListByPair(ctx context.Context, pairID string, limit, offset int32) ([]*entities.Trade, error)
}
