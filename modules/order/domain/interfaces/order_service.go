package interfaces

import (
	"context"

	"booker/modules/order/application/dto"
	"booker/modules/order/domain/entities"

	"github.com/shopspring/decimal"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req *dto.CreateOrderDTO) (*entities.Order, error)
	CancelOrder(ctx context.Context, userID, orderID string) (*entities.Order, error)
	GetOrder(ctx context.Context, userID, orderID string) (*entities.Order, error)
	GetOrderInternal(ctx context.Context, orderID string) (*entities.Order, error)
	ListOrders(ctx context.Context, userID string, req *dto.ListOrdersDTO) ([]*entities.Order, error)
	UpdateOrderFill(
		ctx context.Context,
		orderID string,
		filledQty decimal.Decimal,
		status string,
	) (*entities.Order, error)
}
