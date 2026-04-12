package engine

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

type BookOrder struct {
	ID        string
	UserID    string
	PairID    string
	Side      Side
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Remaining decimal.Decimal
	CreatedAt time.Time
}

type Trade struct {
	ID          string
	PairID      string
	BuyOrderID  string
	SellOrderID string
	BuyerID     string
	SellerID    string
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	ExecutedAt  time.Time
}

func NewTrade(pairID string, buyOrder, sellOrder *BookOrder, price, qty decimal.Decimal) *Trade {
	return &Trade{
		ID:          uuid.New().String(),
		PairID:      pairID,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyerID:     buyOrder.UserID,
		SellerID:    sellOrder.UserID,
		Price:       price,
		Quantity:    qty,
		ExecutedAt:  time.Now(),
	}
}

// Command types for channel-based engine
type CommandType int

const (
	CmdSubmit CommandType = iota
	CmdCancel
	CmdStop
)

type Command struct {
	Type     CommandType
	Order    *BookOrder
	OrderID  string
	ResultCh chan<- Result
}

type Result struct {
	Trades  []*Trade
	Err     error
	OrderID string
}
