package repositories

import (
	"context"

	"booker/modules/wallet/domain"
	"booker/modules/wallet/domain/entities"
	"booker/modules/wallet/domain/interfaces"
	"booker/modules/wallet/infrastructure/gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type walletRepository struct {
	q *gen.Queries
}

func NewWalletRepository(pool *pgxpool.Pool) interfaces.WalletRepository {
	return &walletRepository{q: gen.New(pool)}
}

func (r *walletRepository) GetOrCreate(ctx context.Context, userID, assetID string) (*entities.Wallet, error) {
	// Try insert (ON CONFLICT DO NOTHING), then select
	_, _ = r.q.GetOrCreateWallet(ctx, gen.GetOrCreateWalletParams{
		UserID: userID, AssetID: assetID,
	})

	row, err := r.q.GetWalletByUserAndAsset(ctx, gen.GetWalletByUserAndAssetParams{
		UserID: userID, AssetID: assetID,
	})
	if err != nil {
		return nil, domain.ErrWalletNotFound.Wrap(err)
	}
	return toEntity(row), nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Wallet, error) {
	rows, err := r.q.GetWalletsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	wallets := make([]*entities.Wallet, len(rows))
	for i, row := range rows {
		wallets[i] = toEntity(row)
	}
	return wallets, nil
}

func (r *walletRepository) Deposit(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error) {
	// Ensure wallet exists
	if _, err := r.GetOrCreate(ctx, userID, assetID); err != nil {
		return nil, err
	}

	row, err := r.q.DepositWallet(ctx, gen.DepositWalletParams{
		UserID: userID, AssetID: assetID, Available: amount,
	})
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func (r *walletRepository) Withdraw(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error) {
	row, err := r.q.WithdrawWallet(ctx, gen.WithdrawWalletParams{
		UserID: userID, AssetID: assetID, Available: amount,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInsufficientBalance
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *walletRepository) Hold(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error) {
	row, err := r.q.HoldWallet(ctx, gen.HoldWalletParams{
		UserID: userID, AssetID: assetID, Available: amount,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInsufficientBalance
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *walletRepository) Release(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error) {
	row, err := r.q.ReleaseWallet(ctx, gen.ReleaseWalletParams{
		UserID: userID, AssetID: assetID, Available: amount,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInsufficientLocked
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *walletRepository) Settle(ctx context.Context, userID, assetID string, amount decimal.Decimal) (*entities.Wallet, error) {
	row, err := r.q.SettleWallet(ctx, gen.SettleWalletParams{
		UserID: userID, AssetID: assetID, Locked: amount,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInsufficientLocked
		}
		return nil, err
	}
	return toEntity(row), nil
}

func toEntity(row gen.Wallet) *entities.Wallet {
	return &entities.Wallet{
		ID:        row.ID,
		UserID:    row.UserID,
		AssetID:   row.AssetID,
		Available: row.Available,
		Locked:    row.Locked,
		UpdatedAt: row.UpdatedAt,
	}
}
