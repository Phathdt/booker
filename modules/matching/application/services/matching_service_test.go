package services

import (
	"context"
	"fmt"
	"testing"

	"booker/modules/matching/domain"
	"booker/modules/matching/domain/entities"
	"booker/modules/matching/domain/interfaces"
	"booker/modules/matching/domain/interfaces/mocks"
	"booker/modules/matching/engine"
	pkgnats "booker/pkg/nats"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestService(t *testing.T) (
	map[string]*engine.Engine,
	*mocks.MockTradeRepository,
	*mocks.MockOrderClient,
	*mocks.MockMatchingWalletClient,
	interfaces.MatchingService,
) {
	ctx := context.Background()

	eng := engine.NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	t.Cleanup(func() { eng.Stop() })

	engines := map[string]*engine.Engine{"BTC_USDT": eng}
	tradeRepo := mocks.NewMockTradeRepository(t)
	orderClient := mocks.NewMockOrderClient(t)
	walletClient := mocks.NewMockMatchingWalletClient(t)

	pairs := map[string]*PairInfo{
		"BTC_USDT": {ID: "BTC_USDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}

	svc := NewMatchingService(engines, tradeRepo, orderClient, walletClient, nil, nil, pairs)
	return engines, tradeRepo, orderClient, walletClient, svc
}

func newBookOrder(id, userID string, side engine.Side, price, qty float64) *engine.BookOrder {
	return &engine.BookOrder{
		ID:        id,
		UserID:    userID,
		PairID:    "BTC_USDT",
		Side:      side,
		Price:     decimal.NewFromFloat(price),
		Quantity:  decimal.NewFromFloat(qty),
		Remaining: decimal.NewFromFloat(qty),
	}
}

func TestSubmitOrder_NoMatch_Rests(t *testing.T) {
	_, _, _, _, svc := setupTestService(t)

	order := newBookOrder("b1", "buyer", engine.SideBuy, 49000, 1)
	trades, err := svc.SubmitOrder(context.Background(), order)

	require.NoError(t, err)
	assert.Empty(t, trades)
}

func TestSubmitOrder_Match_SettlesAll(t *testing.T) {
	engines, tradeRepo, orderClient, walletClient, svc := setupTestService(t)

	// Pre-add a sell order
	sell := newBookOrder("a1", "seller", engine.SideSell, 50000, 0.5)
	engines["BTC_USDT"].Submit(sell)

	quoteAmount := decimal.NewFromFloat(25000) // 50000 * 0.5
	qty := decimal.NewFromFloat(0.5)

	// Expect 4 wallet calls
	walletClient.EXPECT().SettleTrade(mock.Anything, "buyer", "USDT", quoteAmount).Return(nil)
	walletClient.EXPECT().Deposit(mock.Anything, "buyer", "BTC", qty).Return(nil)
	walletClient.EXPECT().SettleTrade(mock.Anything, "seller", "BTC", qty).Return(nil)
	walletClient.EXPECT().Deposit(mock.Anything, "seller", "USDT", quoteAmount).Return(nil)

	// Expect order fill updates
	orderClient.EXPECT().UpdateOrderFill(mock.Anything, mock.Anything, qty, "partial").Return(nil).Times(2)

	// Expect trade persistence
	tradeRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Trade{}, nil)

	buy := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 0.5)
	trades, err := svc.SubmitOrder(context.Background(), buy)

	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.True(t, trades[0].Quantity.Equal(qty))
}

func TestSubmitOrder_PairNotFound(t *testing.T) {
	_, _, _, _, svc := setupTestService(t)

	order := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 1)
	order.PairID = "INVALID"

	trades, err := svc.SubmitOrder(context.Background(), order)
	assert.Nil(t, trades)
	assert.Equal(t, domain.ErrPairEngineNotFound, err)
}

func TestSubmitOrder_SettlementFails_LogsError(t *testing.T) {
	engines, tradeRepo, orderClient, walletClient, svc := setupTestService(t)

	sell := newBookOrder("a1", "seller", engine.SideSell, 50000, 0.5)
	engines["BTC_USDT"].Submit(sell)

	// Settlement fails — should log but not crash
	walletClient.EXPECT().SettleTrade(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("wallet down")).Maybe()
	walletClient.EXPECT().Deposit(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("wallet down")).Maybe()
	orderClient.EXPECT().UpdateOrderFill(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("order-svc down")).Maybe()
	tradeRepo.EXPECT().Create(mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("db error")).Maybe()

	buy := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 0.5)
	trades, err := svc.SubmitOrder(context.Background(), buy)

	require.NoError(t, err) // match itself succeeds even if settlement logs errors
	assert.Len(t, trades, 1)
}

func TestCancelOrder_Success(t *testing.T) {
	engines, _, _, _, svc := setupTestService(t)

	order := newBookOrder("b1", "buyer", engine.SideBuy, 49000, 1)
	engines["BTC_USDT"].Submit(order)

	err := svc.CancelOrder(context.Background(), "BTC_USDT", "b1")
	assert.NoError(t, err)
}

func TestCancelOrder_NotInBook(t *testing.T) {
	_, _, _, _, svc := setupTestService(t)

	err := svc.CancelOrder(context.Background(), "BTC_USDT", "nonexistent")
	assert.Equal(t, domain.ErrOrderNotInBook, err)
}

func TestCancelOrder_PairNotFound(t *testing.T) {
	_, _, _, _, svc := setupTestService(t)

	err := svc.CancelOrder(context.Background(), "INVALID", "b1")
	assert.Equal(t, domain.ErrPairEngineNotFound, err)
}

func TestSubmitOrder_WithNATSPublisher(t *testing.T) {
	ctx := context.Background()

	eng := engine.NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	t.Cleanup(func() { eng.Stop() })

	engines := map[string]*engine.Engine{"BTC_USDT": eng}
	tradeRepo := mocks.NewMockTradeRepository(t)
	orderClient := mocks.NewMockOrderClient(t)
	walletClient := mocks.NewMockMatchingWalletClient(t)

	// Use a mock publisher
	publisher := &mockPublisher{}

	pairs := map[string]*PairInfo{
		"BTC_USDT": {ID: "BTC_USDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}

	svc := NewMatchingService(engines, tradeRepo, orderClient, walletClient, publisher, nil, pairs)

	sell := newBookOrder("a1", "seller", engine.SideSell, 50000, 0.5)
	eng.Submit(sell)

	walletClient.EXPECT().SettleTrade(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	walletClient.EXPECT().Deposit(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	orderClient.EXPECT().UpdateOrderFill(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	tradeRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Trade{}, nil).Maybe()

	buy := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 0.5)
	trades, err := svc.SubmitOrder(ctx, buy)

	require.NoError(t, err)
	assert.Len(t, trades, 1)
	assert.Equal(t, 1, publisher.count)
}

func TestSubmitOrder_WithFailingPublisher(t *testing.T) {
	ctx := context.Background()

	eng := engine.NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	t.Cleanup(func() { eng.Stop() })

	engines := map[string]*engine.Engine{"BTC_USDT": eng}
	tradeRepo := mocks.NewMockTradeRepository(t)
	orderClient := mocks.NewMockOrderClient(t)
	walletClient := mocks.NewMockMatchingWalletClient(t)
	publisher := &failingPublisher{}

	pairs := map[string]*PairInfo{
		"BTC_USDT": {ID: "BTC_USDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}

	svc := NewMatchingService(engines, tradeRepo, orderClient, walletClient, publisher, nil, pairs)

	sell := newBookOrder("a1", "seller", engine.SideSell, 50000, 0.5)
	eng.Submit(sell)

	walletClient.EXPECT().SettleTrade(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	walletClient.EXPECT().Deposit(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	orderClient.EXPECT().UpdateOrderFill(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	tradeRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Trade{}, nil).Maybe()

	buy := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 0.5)
	trades, err := svc.SubmitOrder(ctx, buy)

	require.NoError(t, err)
	assert.Len(t, trades, 1)
}

func TestDetermineOrderStatus_Filled(t *testing.T) {
	assert.Equal(t, "filled", DetermineOrderStatus(decimal.NewFromFloat(1), decimal.NewFromFloat(1)))
}

func TestDetermineOrderStatus_Partial(t *testing.T) {
	assert.Equal(t, "partial", DetermineOrderStatus(decimal.NewFromFloat(0.5), decimal.NewFromFloat(1)))
}

func TestDetermineOrderStatus_OverFilled(t *testing.T) {
	assert.Equal(t, "filled", DetermineOrderStatus(decimal.NewFromFloat(1.5), decimal.NewFromFloat(1)))
}

func TestSettleTrade_PairNotFound(t *testing.T) {
	ctx := context.Background()

	eng := engine.NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	t.Cleanup(func() { eng.Stop() })

	engines := map[string]*engine.Engine{"BTC_USDT": eng}
	tradeRepo := mocks.NewMockTradeRepository(t)
	orderClient := mocks.NewMockOrderClient(t)
	walletClient := mocks.NewMockMatchingWalletClient(t)

	// Service has BTC_USDT engine but pairs map is EMPTY
	// This means trades will be generated but settleTrade won't find the pair
	pairs := map[string]*PairInfo{}

	svc := NewMatchingService(engines, tradeRepo, orderClient, walletClient, nil, nil, pairs)

	// Add a sell order
	sell := newBookOrder("a1", "seller", engine.SideSell, 50000, 0.5)
	eng.Submit(sell)

	// Submit buy order to trigger trade generation
	buy := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 0.5)
	trades, err := svc.SubmitOrder(ctx, buy)

	// SubmitOrder itself should succeed (match is found)
	// But settleTrade will find pair not in map and log error
	assert.NoError(t, err)
	assert.Len(t, trades, 1)
}

func TestSubmitOrder_MultipleMatches(t *testing.T) {
	engines, tradeRepo, orderClient, walletClient, svc := setupTestService(t)

	// Add multiple sell orders at different prices
	sell1 := newBookOrder("s1", "seller1", engine.SideSell, 49900, 0.2)
	sell2 := newBookOrder("s2", "seller2", engine.SideSell, 49950, 0.3)
	engines["BTC_USDT"].Submit(sell1)
	engines["BTC_USDT"].Submit(sell2)

	qty1 := decimal.NewFromFloat(0.2)
	qty2 := decimal.NewFromFloat(0.3)

	// Expect wallet calls for all trades
	walletClient.EXPECT().SettleTrade(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(4)
	walletClient.EXPECT().Deposit(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(4)

	// Expect order fill updates
	orderClient.EXPECT().UpdateOrderFill(mock.Anything, mock.Anything, mock.Anything, "partial").Return(nil).Times(4)

	// Expect trade persistence
	tradeRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Trade{}, nil).Times(2)

	// Buy 0.5 total — should match against both sells
	buy := newBookOrder("b1", "buyer", engine.SideBuy, 50000, 0.5)
	trades, err := svc.SubmitOrder(context.Background(), buy)

	require.NoError(t, err)
	require.Len(t, trades, 2)
	assert.True(t, trades[0].Quantity.Equal(qty1))
	assert.True(t, trades[1].Quantity.Equal(qty2))
}

func TestCancelOrder_ThenSubmitSameID(t *testing.T) {
	engines, _, _, _, svc := setupTestService(t)

	order := newBookOrder("b1", "buyer", engine.SideBuy, 49000, 1)
	engines["BTC_USDT"].Submit(order)

	// Cancel the order
	err := svc.CancelOrder(context.Background(), "BTC_USDT", "b1")
	assert.NoError(t, err)

	// Try to cancel again — should fail
	err = svc.CancelOrder(context.Background(), "BTC_USDT", "b1")
	assert.Equal(t, domain.ErrOrderNotInBook, err)
}

type failingPublisher struct{}

func (p *failingPublisher) PublishTrade(_ context.Context, _ *pkgnats.TradeEvent) error {
	return fmt.Errorf("nats down")
}

// mockPublisher is a simple in-memory publisher for testing.
type mockPublisher struct {
	count int
}

func (p *mockPublisher) PublishTrade(_ context.Context, _ *pkgnats.TradeEvent) error {
	p.count++
	return nil
}
