package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

type Trade struct {
	ID          string
	PairID      string
	BuyOrderID  string
	SellOrderID string
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	BuyerID     string
	SellerID    string
	ExecutedAt  time.Time
}
