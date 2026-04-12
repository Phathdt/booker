package interfaces

import (
	"context"

	"booker/modules/order/domain/entities"
)

type MatchingClient interface {
	SubmitOrder(ctx context.Context, order *entities.Order) error
	CancelOrder(ctx context.Context, pairID, orderID string) error
}
