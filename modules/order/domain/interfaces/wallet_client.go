package interfaces

import (
	"context"

	"github.com/shopspring/decimal"
)

type WalletClient interface {
	HoldBalance(ctx context.Context, userID, assetID string, amount decimal.Decimal) error
	ReleaseBalance(ctx context.Context, userID, assetID string, amount decimal.Decimal) error
}
