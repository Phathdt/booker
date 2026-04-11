package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

// Wallet represents a user's balance for a specific asset.
type Wallet struct {
	ID        string
	UserID    string
	AssetID   string
	Available decimal.Decimal
	Locked    decimal.Decimal
	UpdatedAt time.Time
}

// Total returns available + locked balance.
func (w *Wallet) Total() decimal.Decimal {
	return w.Available.Add(w.Locked)
}
