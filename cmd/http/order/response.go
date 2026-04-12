package order

import (
	"time"

	"booker/modules/order/application/dto"
	"booker/modules/order/domain/entities"
)

func toOrderResponse(o *entities.Order) dto.OrderResponse {
	return dto.OrderResponse{
		ID:        o.ID,
		UserID:    o.UserID,
		PairID:    o.PairID,
		Side:      o.Side,
		Type:      o.Type,
		Price:     o.Price.String(),
		Quantity:  o.Quantity.String(),
		FilledQty: o.FilledQty.String(),
		Status:    o.Status,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
	}
}

func toOrderListResponse(orders []*entities.Order) dto.OrderListResponse {
	items := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		items[i] = toOrderResponse(o)
	}
	return dto.OrderListResponse{Orders: items}
}
