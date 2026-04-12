package services

import (
	"context"
	"log/slog"
	"time"

	"booker/modules/wallet/domain"
	"booker/modules/wallet/domain/entities"
	"booker/modules/wallet/domain/interfaces"
	pkgnats "booker/pkg/nats"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type walletService struct {
	repo            interfaces.WalletRepository
	walletPublisher pkgnats.WalletPublisher
}

// NewWalletService creates a new WalletService.
func NewWalletService(repo interfaces.WalletRepository, walletPublisher pkgnats.WalletPublisher) interfaces.WalletService {
	return &walletService{repo: repo, walletPublisher: walletPublisher}
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
	wallet, err := s.repo.Deposit(ctx, userID, assetID, amount)
	if err != nil {
		return nil, err
	}
	s.publishWalletEvent(ctx, userID, assetID, amount, "deposit")
	return wallet, nil
}

func (s *walletService) Withdraw(
	ctx context.Context,
	userID, assetID string,
	amount decimal.Decimal,
) (*entities.Wallet, error) {
	if err := validateAmount(amount); err != nil {
		return nil, err
	}
	wallet, err := s.repo.Withdraw(ctx, userID, assetID, amount)
	if err != nil {
		return nil, err
	}
	s.publishWalletEvent(ctx, userID, assetID, amount, "withdrawal")
	return wallet, nil
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

func (s *walletService) publishWalletEvent(ctx context.Context, userID, asset string, amount decimal.Decimal, action string) {
	if s.walletPublisher == nil {
		return
	}
	event := &pkgnats.WalletEvent{
		UserID:    userID,
		Asset:     asset,
		Amount:    amount.String(),
		Action:    action,
		TxID:      uuid.NewString(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	if err := s.walletPublisher.PublishWalletAction(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to publish wallet event",
			"user_id", userID, "action", action, "error", err.Error())
	}
}
