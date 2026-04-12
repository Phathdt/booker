package interfaces

import (
	"context"

	"github.com/shopspring/decimal"
)

type OrderClient interface {
	UpdateOrderFill(ctx context.Context, orderID string, filledQty decimal.Decimal, status string) error
}
