package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID        string
	UserID    string
	PairID    string
	Side      string // "buy" or "sell"
	Type      string // "limit"
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	FilledQty decimal.Decimal
	Status    string // "new", "partial", "filled", "cancelled"
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o *Order) RemainingQty() decimal.Decimal {
	return o.Quantity.Sub(o.FilledQty)
}

func (o *Order) IsCancellable() bool {
	return o.Status == "new" || o.Status == "partial"
}
