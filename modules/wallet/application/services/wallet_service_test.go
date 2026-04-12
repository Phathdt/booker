package services

import (
	"context"
	"testing"

	"booker/modules/wallet/domain"
	"booker/modules/wallet/domain/entities"
	"booker/modules/wallet/domain/interfaces/mocks"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletService_Deposit_Success(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	amount := decimal.NewFromFloat(100.5)
	repo.EXPECT().Deposit(mock.Anything, "user-1", "USDT", amount).
		Return(&entities.Wallet{ID: "w-1", Available: amount}, nil)

	w, err := svc.Deposit(context.Background(), "user-1", "USDT", amount)
	assert.NoError(t, err)
	assert.True(t, w.Available.Equal(amount))
}

func TestWalletService_Deposit_InvalidAmount(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	w, err := svc.Deposit(context.Background(), "user-1", "USDT", decimal.Zero)
	assert.Nil(t, w)
	assert.Equal(t, domain.ErrInvalidAmount, err)
}

func TestWalletService_Deposit_NegativeAmount(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	w, err := svc.Deposit(context.Background(), "user-1", "USDT", decimal.NewFromFloat(-10))
	assert.Nil(t, w)
	assert.Equal(t, domain.ErrInvalidAmount, err)
}

func TestWalletService_Withdraw_Success(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	amount := decimal.NewFromFloat(50)
	repo.EXPECT().Withdraw(mock.Anything, "user-1", "USDT", amount).
		Return(&entities.Wallet{ID: "w-1", Available: decimal.NewFromFloat(950)}, nil)

	w, err := svc.Withdraw(context.Background(), "user-1", "USDT", amount)
	assert.NoError(t, err)
	assert.True(t, w.Available.Equal(decimal.NewFromFloat(950)))
}

func TestWalletService_Withdraw_InvalidAmount(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	w, err := svc.Withdraw(context.Background(), "user-1", "USDT", decimal.Zero)
	assert.Nil(t, w)
	assert.Equal(t, domain.ErrInvalidAmount, err)
}

func TestWalletService_HoldBalance_Success(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	amount := decimal.NewFromFloat(100)
	repo.EXPECT().Hold(mock.Anything, "user-1", "USDT", amount).
		Return(&entities.Wallet{Available: decimal.NewFromFloat(900), Locked: amount}, nil)

	w, err := svc.HoldBalance(context.Background(), "user-1", "USDT", amount)
	assert.NoError(t, err)
	assert.True(t, w.Locked.Equal(amount))
}

func TestWalletService_HoldBalance_InvalidAmount(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	w, err := svc.HoldBalance(context.Background(), "user-1", "USDT", decimal.NewFromFloat(-1))
	assert.Nil(t, w)
	assert.Equal(t, domain.ErrInvalidAmount, err)
}

func TestWalletService_ReleaseBalance_Success(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	amount := decimal.NewFromFloat(50)
	repo.EXPECT().Release(mock.Anything, "user-1", "USDT", amount).
		Return(&entities.Wallet{Available: decimal.NewFromFloat(950), Locked: decimal.NewFromFloat(50)}, nil)

	w, err := svc.ReleaseBalance(context.Background(), "user-1", "USDT", amount)
	assert.NoError(t, err)
	assert.NotNil(t, w)
}

func TestWalletService_SettleTrade_Success(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	amount := decimal.NewFromFloat(100)
	repo.EXPECT().Settle(mock.Anything, "user-1", "USDT", amount).
		Return(&entities.Wallet{Locked: decimal.Zero}, nil)

	w, err := svc.SettleTrade(context.Background(), "user-1", "USDT", amount)
	assert.NoError(t, err)
	assert.True(t, w.Locked.Equal(decimal.Zero))
}

func TestWalletService_GetBalance(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	repo.EXPECT().GetOrCreate(mock.Anything, "user-1", "BTC").
		Return(&entities.Wallet{UserID: "user-1", AssetID: "BTC"}, nil)

	w, err := svc.GetBalance(context.Background(), "user-1", "BTC")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", w.AssetID)
}

func TestWalletService_GetBalances(t *testing.T) {
	repo := mocks.NewMockWalletRepository(t)
	svc := NewWalletService(repo, nil)

	repo.EXPECT().GetByUserID(mock.Anything, "user-1").
		Return([]*entities.Wallet{{AssetID: "BTC"}, {AssetID: "USDT"}}, nil)

	wallets, err := svc.GetBalances(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Len(t, wallets, 2)
}
