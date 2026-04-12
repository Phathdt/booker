package services

import (
	"context"

	"booker/modules/wallet/domain"
	"booker/modules/wallet/domain/entities"
	"booker/modules/wallet/domain/interfaces"

	"github.com/shopspring/decimal"
)

type walletService struct {
	repo interfaces.WalletRepository
}

// NewWalletService creates a new WalletService.
func NewWalletService(repo interfaces.WalletRepository) interfaces.WalletService {
	return &walletService{repo: repo}
}

func (s *walletService) GetBalance(ctx context.Context, userID, assetID string) (*entities.Wallet, error) {
	return s.repo.GetOrCreate(ctx, userID, assetID)
}

func (s *walletService) GetBalances(ctx context.Context, userID string) ([]*entities.Wallet, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *walletService) Deposit(
	ctx context.Context,
	userID, assetID string,
	amount decimal.Decimal,
) (*entities.Wallet, error) {
	if err := validateAmount(amount); err != nil {
		return nil, err
	}
	return s.repo.Deposit(ctx, userID, assetID, amount)
}

func (s *walletService) Withdraw(
	ctx context.Context,
	userID, assetID string,
	amount decimal.Decimal,
) (*entities.Wallet, error) {
	if err := validateAmount(amount); err != nil {
		return nil, err
	}
	return s.repo.Withdraw(ctx, userID, assetID, amount)
}

func (s *walletService) HoldBalance(
	ctx context.Context,
	userID, assetID string,
	amount decimal.Decimal,
) (*entities.Wallet, error) {
	if err := validateAmount(amount); err != nil {
		return nil, err
	}
	return s.repo.Hold(ctx, userID, assetID, amount)
}

func (s *walletService) ReleaseBalance(
	ctx context.Context,
	userID, assetID string,
	amount decimal.Decimal,
) (*entities.Wallet, error) {
	if err := validateAmount(amount); err != nil {
		return nil, err
	}
	return s.repo.Release(ctx, userID, assetID, amount)
}

func (s *walletService) SettleTrade(
	ctx context.Context,
	userID, assetID string,
	amount decimal.Decimal,
) (*entities.Wallet, error) {
	if err := validateAmount(amount); err != nil {
		return nil, err
	}
	return s.repo.Settle(ctx, userID, assetID, amount)
}

func validateAmount(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return domain.ErrInvalidAmount
	}
	return nil
}
