package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"booker/modules/order/application/dto"
	"booker/modules/order/domain"
	"booker/modules/order/domain/entities"
	"booker/modules/order/domain/interfaces"
	"booker/modules/order/domain/interfaces/mocks"
	pkgnats "booker/pkg/nats"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testPair = &interfaces.TradingPair{
	ID:         "BTC_USDT",
	BaseAsset:  "BTC",
	QuoteAsset: "USDT",
	Status:     "active",
	MinQty:     decimal.NewFromFloat(0.00001),
	TickSize:   decimal.NewFromFloat(0.01),
}

func newTestService(t *testing.T) (*mocks.MockOrderRepository, *mocks.MockWalletClient, interfaces.OrderService) {
	repo := mocks.NewMockOrderRepository(t)
	wallet := mocks.NewMockWalletClient(t)
	matching := mocks.NewMockMatchingClient(t)
	// Default: matching client submit always succeeds (fire-and-forget)
	matching.EXPECT().SubmitOrder(mock.Anything, mock.Anything).Return(nil).Maybe()
	matching.EXPECT().CancelOrder(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	svc := NewOrderService(repo, wallet, matching, nil)
	return repo, wallet, svc
}

func validCreateDTO() *dto.CreateOrderDTO {
	return &dto.CreateOrderDTO{
		PairID:   "BTC_USDT",
		Side:     "buy",
		Type:     "limit",
		Price:    decimal.NewFromFloat(50000),
		Quantity: decimal.NewFromFloat(0.5),
	}
}

// --- CreateOrder tests ---

func TestCreateOrder_Success_Buy(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	req := validCreateDTO()
	holdAmount := req.Price.Mul(req.Quantity) // 25000

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", holdAmount).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: req.Price, Quantity: req.Quantity, Status: "new",
	}, nil)

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "order-1", order.ID)
	assert.Equal(t, "new", order.Status)
}

func TestCreateOrder_Success_Sell(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	req := validCreateDTO()
	req.Side = "sell"

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "BTC", req.Quantity).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Order{
		ID: "order-2", Side: "sell", Status: "new",
	}, nil)

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "sell", order.Side)
}

func TestCreateOrder_InvalidPrice(t *testing.T) {
	_, _, svc := newTestService(t)

	req := validCreateDTO()
	req.Price = decimal.Zero

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInvalidPrice, err)
}

func TestCreateOrder_InvalidQuantity(t *testing.T) {
	_, _, svc := newTestService(t)

	req := validCreateDTO()
	req.Quantity = decimal.NewFromFloat(-1)

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInvalidQuantity, err)
}

func TestCreateOrder_InvalidSide(t *testing.T) {
	_, _, svc := newTestService(t)

	req := validCreateDTO()
	req.Side = "invalid"

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInvalidSide, err)
}

func TestCreateOrder_InvalidOrderType(t *testing.T) {
	_, _, svc := newTestService(t)

	req := validCreateDTO()
	req.Type = "market"

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInvalidOrderType, err)
}

func TestCreateOrder_PairNotFound(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetTradingPair(mock.Anything, "INVALID").Return(nil, domain.ErrPairNotFound)

	req := validCreateDTO()
	req.PairID = "INVALID"

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrPairNotFound, err)
}

func TestCreateOrder_PairNotActive(t *testing.T) {
	repo, _, svc := newTestService(t)

	inactivePair := &interfaces.TradingPair{
		ID: "BTC_USDT", BaseAsset: "BTC", QuoteAsset: "USDT",
		Status: "inactive", MinQty: decimal.NewFromFloat(0.00001), TickSize: decimal.NewFromFloat(0.01),
	}
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(inactivePair, nil)

	order, err := svc.CreateOrder(context.Background(), "user-1", validCreateDTO())
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrPairNotActive, err)
}

func TestCreateOrder_BelowMinQty(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)

	req := validCreateDTO()
	req.Quantity = decimal.NewFromFloat(0.000001) // below min_qty 0.00001

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrBelowMinQty, err)
}

func TestCreateOrder_InvalidTickSize(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)

	req := validCreateDTO()
	req.Price = decimal.NewFromFloat(50000.005) // not multiple of 0.01

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInvalidTickSize, err)
}

func TestCreateOrder_InsufficientBalance(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", mock.Anything).
		Return(domain.ErrInsufficientBalance)

	order, err := svc.CreateOrder(context.Background(), "user-1", validCreateDTO())
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInsufficientBalance, err)
}

func TestCreateOrder_DBError_Rollback(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	req := validCreateDTO()
	holdAmount := req.Price.Mul(req.Quantity)

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", holdAmount).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", holdAmount).Return(nil)

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Error(t, err)
}

func TestCreateOrder_DBError_RollbackFails(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	req := validCreateDTO()
	holdAmount := req.Price.Mul(req.Quantity)

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", holdAmount).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", holdAmount).Return(fmt.Errorf("release failed"))

	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.Nil(t, order)
	assert.Error(t, err)
}

// --- CancelOrder tests ---

func TestCancelOrder_Success_WalletFirst(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}
	// Buy-side release: price * remainingQty
	releaseAmount := order.Price.Mul(order.RemainingQty())

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(&entities.Order{
		ID: "order-1", Status: "cancelled",
	}, nil)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", result.Status)
}

func TestCancelOrder_SellSide_ReleaseAmount(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "sell",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		FilledQty: decimal.NewFromFloat(0.3), Status: "partial",
	}
	// Sell-side release: just remainingQty on base asset
	remainingQty := order.RemainingQty() // 0.7

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "BTC", remainingQty).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(&entities.Order{
		ID: "order-1", Status: "cancelled",
	}, nil)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", result.Status)
}

func TestCancelOrder_WalletFail_OrderStaysActive(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}
	releaseAmount := order.Price.Mul(order.RemainingQty())

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).
		Return(fmt.Errorf("wallet unavailable"))

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestCancelOrder_NotFound(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").
		Return(nil, domain.ErrOrderNotFound)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrOrderNotFound, err)
}

func TestCancelOrder_BuySide_PartiallyFilled(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		FilledQty: decimal.NewFromFloat(0.4), Status: "partial",
	}
	// Buy-side: price * remainingQty = 50000 * 0.6 = 30000
	releaseAmount := order.Price.Mul(order.RemainingQty())

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(&entities.Order{
		ID: "order-1", Status: "cancelled",
	}, nil)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", result.Status)
}

func TestCancelOrder_NotCancellable(t *testing.T) {
	repo, _, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		FilledQty: decimal.NewFromFloat(1), Status: "filled",
	}

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrOrderNotCancellable, err)
}

func TestCancelOrder_DBFail_CompensatingHold(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}
	releaseAmount := order.Price.Mul(order.RemainingQty())

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(nil, fmt.Errorf("db error"))
	// Compensating hold should be called
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrOrderNotCancellable, err)
}

func TestCancelOrder_DBFail_CompensatingHoldAlsoFails(t *testing.T) {
	repo, wallet, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}
	releaseAmount := order.Price.Mul(order.RemainingQty())

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(nil, fmt.Errorf("db error"))
	// Compensating hold also fails — CRITICAL log
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(fmt.Errorf("hold failed"))

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrOrderNotCancellable, err)
}

func TestCancelOrder_PairNotFound(t *testing.T) {
	repo, _, svc := newTestService(t)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "INVALID", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "INVALID").Return(nil, domain.ErrPairNotFound)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrPairNotFound, err)
}

// --- GetOrder tests ---

func TestGetOrder_Success(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").
		Return(&entities.Order{ID: "order-1"}, nil)

	order, err := svc.GetOrder(context.Background(), "user-1", "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "order-1", order.ID)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").
		Return(nil, domain.ErrOrderNotFound)

	order, err := svc.GetOrder(context.Background(), "user-1", "order-1")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrOrderNotFound, err)
}

func TestGetOrder_EmptyUserID_ReturnsNotFound(t *testing.T) {
	_, _, svc := newTestService(t)

	order, err := svc.GetOrder(context.Background(), "", "order-1")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrOrderNotFound, err)
}

func TestGetOrderInternal_Success(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1"}, nil)

	order, err := svc.GetOrderInternal(context.Background(), "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "order-1", order.ID)
}

func TestGetOrderInternal_NotFound(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(nil, domain.ErrOrderNotFound)

	order, err := svc.GetOrderInternal(context.Background(), "order-1")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrOrderNotFound, err)
}

// --- ListOrders tests ---

func TestListOrders_Success(t *testing.T) {
	repo, _, svc := newTestService(t)

	pairID := "BTC_USDT"
	repo.EXPECT().List(mock.Anything, "user-1", &pairID, (*string)(nil), int32(20), int32(0)).
		Return([]*entities.Order{{ID: "o1"}, {ID: "o2"}}, nil)

	orders, err := svc.ListOrders(context.Background(), "user-1", &dto.ListOrdersDTO{
		PairID: "BTC_USDT",
	})
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestListOrders_DefaultLimit(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().List(mock.Anything, "user-1", (*string)(nil), (*string)(nil), int32(20), int32(0)).
		Return([]*entities.Order{}, nil)

	orders, err := svc.ListOrders(context.Background(), "user-1", &dto.ListOrdersDTO{})
	assert.NoError(t, err)
	assert.Empty(t, orders)
}

func TestListOrders_CustomLimit(t *testing.T) {
	repo, _, svc := newTestService(t)

	status := "new"
	repo.EXPECT().List(mock.Anything, "user-1", (*string)(nil), &status, int32(50), int32(10)).
		Return([]*entities.Order{{ID: "o1"}}, nil)

	orders, err := svc.ListOrders(context.Background(), "user-1", &dto.ListOrdersDTO{
		Status: "new",
		Limit:  50,
		Offset: 10,
	})
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
}

func TestListOrders_RepoError(t *testing.T) {
	repo, _, svc := newTestService(t)

	repo.EXPECT().List(mock.Anything, "user-1", (*string)(nil), (*string)(nil), int32(20), int32(0)).
		Return(nil, fmt.Errorf("db error"))

	orders, err := svc.ListOrders(context.Background(), "user-1", &dto.ListOrdersDTO{})
	assert.Nil(t, orders)
	assert.Error(t, err)
}

// --- UpdateOrderFill tests ---

func TestUpdateOrderFill_Success(t *testing.T) {
	repo, _, svc := newTestService(t)

	filledQty := decimal.NewFromFloat(0.5)
	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1", FilledQty: decimal.NewFromFloat(0.3), Status: "partial"}, nil)
	repo.EXPECT().UpdateFilledQty(mock.Anything, "order-1", filledQty, "partial").
		Return(&entities.Order{ID: "order-1", FilledQty: filledQty, Status: "partial"}, nil)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "partial")
	assert.NoError(t, err)
	assert.Equal(t, "partial", order.Status)
}

func TestUpdateOrderFill_InvalidFillQty(t *testing.T) {
	_, _, svc := newTestService(t)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", decimal.Zero, "partial")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrInvalidFillQty, err)
}

func TestUpdateOrderFill_InvalidStatus(t *testing.T) {
	_, _, svc := newTestService(t)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", decimal.NewFromFloat(1), "cancelled")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrOrderNotFillable, err)
}

func TestUpdateOrderFill_OrderNotFillable(t *testing.T) {
	repo, _, svc := newTestService(t)

	filledQty := decimal.NewFromFloat(0.5)
	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1", FilledQty: decimal.NewFromFloat(0.3), Status: "partial"}, nil)
	repo.EXPECT().UpdateFilledQty(mock.Anything, "order-1", filledQty, "filled").
		Return(nil, domain.ErrOrderNotFillable)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "filled")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrOrderNotFillable, err)
}

func TestUpdateOrderFill_BackwardFill(t *testing.T) {
	repo, _, svc := newTestService(t)

	filledQty := decimal.NewFromFloat(0.2)
	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1", FilledQty: decimal.NewFromFloat(0.5), Status: "partial"}, nil)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "partial")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrFillQtyBackward, err)
}

func TestUpdateOrderFill_OrderNotFound(t *testing.T) {
	repo, _, svc := newTestService(t)

	filledQty := decimal.NewFromFloat(0.5)
	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(nil, domain.ErrOrderNotFound)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "partial")
	assert.Nil(t, order)
	assert.Equal(t, domain.ErrOrderNotFound, err)
}

func TestUpdateOrderFill_EqualFill(t *testing.T) {
	repo, _, svc := newTestService(t)

	filledQty := decimal.NewFromFloat(0.5)
	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1", FilledQty: decimal.NewFromFloat(0.5), Status: "partial"}, nil)
	repo.EXPECT().UpdateFilledQty(mock.Anything, "order-1", filledQty, "filled").
		Return(&entities.Order{ID: "order-1", FilledQty: filledQty, Status: "filled"}, nil)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "filled")
	assert.NoError(t, err)
	assert.Equal(t, "filled", order.Status)
}

// --- CreateOrder with matching client failure ---

func TestCreateOrder_SubmitOrderFails(t *testing.T) {
	repo, wallet, _ := newTestService(t)

	// Use a separate matching mock that fails
	matchingFail := mocks.NewMockMatchingClient(t)
	matchingFail.EXPECT().SubmitOrder(mock.Anything, mock.Anything).
		Return(fmt.Errorf("matching engine down"))

	svc := NewOrderService(repo, wallet, matchingFail, nil)

	req := validCreateDTO()
	holdAmount := req.Price.Mul(req.Quantity)

	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().HoldBalance(mock.Anything, "user-1", "USDT", holdAmount).Return(nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(&entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: req.Price, Quantity: req.Quantity, Status: "new",
	}, nil)

	// Should still return successfully (fire-and-forget)
	order, err := svc.CreateOrder(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Equal(t, "order-1", order.ID)
}

// --- CancelOrder with matching client failure ---

func TestCancelOrder_MatchingClientFails(t *testing.T) {
	repo, wallet, _ := newTestService(t)

	// Use a separate matching mock that fails
	matchingFail := mocks.NewMockMatchingClient(t)
	matchingFail.EXPECT().CancelOrder(mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("matching engine down"))

	svc := NewOrderService(repo, wallet, matchingFail, nil)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}
	releaseAmount := order.Price.Mul(order.RemainingQty())

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(&entities.Order{
		ID: "order-1", Status: "cancelled",
	}, nil)

	// Should still return successfully (fire-and-forget)
	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", result.Status)
}

// --- PublishOrderEvent with publisher ---

type testPublisher struct {
	published bool
	err       error
}

func (p *testPublisher) PublishOrderUpdate(ctx context.Context, event *pkgnats.OrderEvent) error {
	if p.err != nil {
		return p.err
	}
	p.published = true
	return nil
}

func TestCancelOrder_WithPublisher(t *testing.T) {
	repo, wallet, _ := newTestService(t)

	publisher := &testPublisher{}
	svc := NewOrderService(repo, wallet, nil, publisher)

	order := &entities.Order{
		ID: "order-1", UserID: "user-1", PairID: "BTC_USDT", Side: "buy",
		Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(0.5),
		FilledQty: decimal.Zero, Status: "new",
	}
	releaseAmount := order.Price.Mul(order.RemainingQty())

	cancelledOrder := &entities.Order{
		ID: "order-1", UserID: "user-1", Status: "cancelled",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	repo.EXPECT().GetByIDAndUser(mock.Anything, "order-1", "user-1").Return(order, nil)
	repo.EXPECT().GetTradingPair(mock.Anything, "BTC_USDT").Return(testPair, nil)
	wallet.EXPECT().ReleaseBalance(mock.Anything, "user-1", "USDT", releaseAmount).Return(nil)
	repo.EXPECT().Cancel(mock.Anything, "order-1", "user-1").Return(cancelledOrder, nil)

	result, err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", result.Status)
	assert.True(t, publisher.published)
}

func TestUpdateOrderFill_WithPublisher(t *testing.T) {
	repo, _, _ := newTestService(t)

	publisher := &testPublisher{}
	svc := NewOrderService(repo, nil, nil, publisher)

	filledQty := decimal.NewFromFloat(0.5)
	updatedOrder := &entities.Order{
		ID: "order-1", FilledQty: filledQty, Status: "partial",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1", FilledQty: decimal.NewFromFloat(0.3), Status: "partial"}, nil)
	repo.EXPECT().UpdateFilledQty(mock.Anything, "order-1", filledQty, "partial").
		Return(updatedOrder, nil)

	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "partial")
	assert.NoError(t, err)
	assert.Equal(t, "partial", order.Status)
	assert.True(t, publisher.published)
}

func TestPublishOrderEvent_FailedPublish(t *testing.T) {
	repo, _, _ := newTestService(t)

	publisher := &testPublisher{err: fmt.Errorf("nats down")}
	svc := NewOrderService(repo, nil, nil, publisher)

	filledQty := decimal.NewFromFloat(0.5)
	updatedOrder := &entities.Order{
		ID: "order-1", FilledQty: filledQty, Status: "partial",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	repo.EXPECT().GetByID(mock.Anything, "order-1").
		Return(&entities.Order{ID: "order-1", FilledQty: decimal.NewFromFloat(0.3), Status: "partial"}, nil)
	repo.EXPECT().UpdateFilledQty(mock.Anything, "order-1", filledQty, "partial").
		Return(updatedOrder, nil)

	// Should not return error even if publish fails (best-effort)
	order, err := svc.UpdateOrderFill(context.Background(), "order-1", filledQty, "partial")
	assert.NoError(t, err)
	assert.Equal(t, "partial", order.Status)
}
