package interfaces

import (
	"context"

	"booker/modules/wallet/domain/entities"

	"github.com/shopspring/decimal"
)

// WalletRepository defines data access for wallets.
type WalletRepository interface {
	// GetByUserAndAsset returns wallet for a user+asset pair. Creates if not exists.
	GetOrCreate(ctx context.Context, userID, assetID string) (*entities.Wallet, error)
	// GetByUserID returns all wallets for a user.
	GetByUserID(ctx context.Context, userID string) ([]*entities.Wallet, error)
	// Deposit adds amount to available balance.
	Deposit(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	// Withdraw subtracts amount from available balance.
	Withdraw(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	// Hold moves amount from available to locked (for placing orders).
	Hold(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	// Release moves amount from locked back to available (for cancelling orders).
	Release(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
	// Settle deducts from locked (trade executed).
	Settle(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error)
}
