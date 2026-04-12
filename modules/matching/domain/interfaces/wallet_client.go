package interfaces

import (
	"context"

	"github.com/shopspring/decimal"
)

type WalletClient interface {
	SettleTrade(ctx context.Context, userID, assetID string, amount decimal.Decimal) error
	Deposit(ctx context.Context, userID, assetID string, amount decimal.Decimal) error
}
