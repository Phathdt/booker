package interfaces

import (
	"context"

	"booker/modules/wallet/domain/entities"

	"github.com/shopspring/decimal"
)

// WalletService defines business logic for wallets.
type WalletService interface {
	GetBalance(ctx context.Context, userID, assetID string) (*entities.Wallet, error)
	GetBalances(ctx context.Context, userID string) ([]*entities.Wallet, error)
	Deposit(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	Withdraw(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	HoldBalance(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	ReleaseBalance(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	SettleTrade(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
}
