package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"booker/modules/notification/domain/entities"
	"booker/modules/notification/domain/interfaces/mocks"
	pkgnats "booker/pkg/nats"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestEventHandler(t *testing.T) (*mocks.MockNotificationService, *EventHandler) {
	svc := mocks.NewMockNotificationService(t)
	handler := NewEventHandler(svc)
	return svc, handler
}

// --- Trade Event Tests ---

func TestHandleTradeEvent_CreatesTwoNotifications(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.TradeEvent{
		TradeID:     "trade-1",
		PairID:      "BTC-USDT",
		BuyOrderID:  "buy-1",
		SellOrderID: "sell-1",
		Price:       "50000",
		Quantity:    "0.5",
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		ExecutedAt:  "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	buyerCalled := false
	sellerCalled := false

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		if n.UserID == "buyer-1" {
			buyerCalled = true
			assert.Equal(t, entities.TypeTradeExecuted, n.Type)
			assert.Equal(t, "Trade Executed", n.Title)
			assert.Contains(t, n.Body, "bought")
			assert.Contains(t, n.Body, "0.5")
			assert.Contains(t, n.Body, "BTC-USDT")
			assert.Equal(t, "trade-1", n.Metadata["trade_id"])
			return true
		}
		return false
	})).Return(true, nil).Once()

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		if n.UserID == "seller-1" {
			sellerCalled = true
			assert.Equal(t, entities.TypeTradeExecuted, n.Type)
			assert.Contains(t, n.Body, "sold")
			assert.Equal(t, "trade-1", n.Metadata["trade_id"])
			return true
		}
		return false
	})).Return(true, nil).Once()

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
	assert.True(t, buyerCalled, "Buyer notification not called")
	assert.True(t, sellerCalled, "Seller notification not called")
}

func TestHandleTradeEvent_MarshalError(t *testing.T) {
	_, handler := newTestEventHandler(t)

	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    []byte("invalid json"),
	}

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal trade event")
}

// --- Order Event Tests ---

func TestHandleOrderEvent_Filled(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.OrderEvent{
		OrderID:   "order-1",
		UserID:    "user-1",
		PairID:    "BTC-USDT",
		Side:      "buy",
		Price:     "50000",
		Quantity:  "1",
		FilledQty: "1",
		Status:    "filled",
		UpdatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    data,
	}

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "user-1" &&
			n.Type == entities.TypeOrderFilled &&
			n.Title == "Order Filled" &&
			n.Metadata["order_id"] == "order-1"
	})).Return(true, nil).Once()

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleOrderEvent_Cancelled(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.OrderEvent{
		OrderID:   "order-2",
		UserID:    "user-2",
		PairID:    "ETH-USDT",
		Side:      "sell",
		Price:     "3000",
		Quantity:  "10",
		FilledQty: "0",
		Status:    "cancelled",
		UpdatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "orders.user-2.cancelled",
		Data:    data,
	}

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "user-2" &&
			n.Type == entities.TypeOrderCancelled &&
			n.Title == "Order Cancelled" &&
			n.Metadata["order_id"] == "order-2"
	})).Return(true, nil).Once()

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleOrderEvent_IgnoresUnknownStatus(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.OrderEvent{
		OrderID: "order-3",
		UserID:  "user-3",
		Status:  "pending",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "orders.user-3.pending",
		Data:    data,
	}

	// Should not call CreateNotification for unknown status
	svc.AssertNotCalled(t, "CreateNotification")

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleOrderEvent_MarshalError(t *testing.T) {
	_, handler := newTestEventHandler(t)

	msg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    []byte("invalid json"),
	}

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal order event")
}

// --- Wallet Event Tests ---

func TestHandleWalletEvent_Deposit(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.WalletEvent{
		UserID:    "user-1",
		Asset:     "BTC",
		Amount:    "0.5",
		Action:    "deposit",
		TxID:      "tx-1",
		CreatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "wallets.user-1.deposit",
		Data:    data,
	}

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "user-1" &&
			n.Type == entities.TypeDepositConfirmed &&
			n.Title == "Deposit Confirmed" &&
			n.Metadata["asset"] == "BTC" &&
			n.Metadata["amount"] == "0.5" &&
			n.Metadata["tx_id"] == "tx-1"
	})).Return(true, nil).Once()

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleWalletEvent_Withdrawal(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.WalletEvent{
		UserID:    "user-2",
		Asset:     "USDT",
		Amount:    "1000",
		Action:    "withdrawal",
		TxID:      "tx-2",
		CreatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "wallets.user-2.withdrawal",
		Data:    data,
	}

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "user-2" &&
			n.Type == entities.TypeWithdrawalConfirmed &&
			n.Title == "Withdrawal Confirmed" &&
			n.Metadata["asset"] == "USDT" &&
			n.Metadata["amount"] == "1000"
	})).Return(true, nil).Once()

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleWalletEvent_IgnoresUnknownAction(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.WalletEvent{
		UserID:    "user-3",
		Asset:     "BTC",
		Amount:    "1",
		Action:    "unknown",
		TxID:      "tx-3",
		CreatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "wallets.user-3.unknown",
		Data:    data,
	}

	// Should not call CreateNotification for unknown action
	svc.AssertNotCalled(t, "CreateNotification")

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleWalletEvent_MarshalError(t *testing.T) {
	_, handler := newTestEventHandler(t)

	msg := &nats.Msg{
		Subject: "wallets.user-1.deposit",
		Data:    []byte("invalid json"),
	}

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal wallet event")
}

// --- Router Tests ---

func TestHandle_RoutesTradeEvent(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.TradeEvent{
		TradeID:  "t1",
		PairID:   "BTC-USDT",
		BuyerID:  "buyer",
		SellerID: "seller",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(event)

	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Times(2)

	msg := &nats.Msg{Subject: "trades.BTC-USDT.executed", Data: data}
	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandle_RoutesOrderEvent(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.OrderEvent{
		OrderID:  "o1",
		UserID:   "user",
		Status:   "filled",
		Side:     "buy",
		Quantity: "1",
		PairID:   "BTC-USDT",
	}
	data, _ := json.Marshal(event)

	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()

	msg := &nats.Msg{Subject: "orders.user.filled", Data: data}
	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandle_RoutesWalletEvent(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.WalletEvent{
		UserID: "user",
		Asset:  "BTC",
		Amount: "1",
		Action: "deposit",
		TxID:   "tx1",
	}
	data, _ := json.Marshal(event)

	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()

	msg := &nats.Msg{Subject: "wallets.user.deposit", Data: data}
	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandle_UnknownSubject(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	msg := &nats.Msg{Subject: "unknown.subject", Data: []byte("data")}

	// Should silently ignore unknown subjects
	svc.AssertNotCalled(t, "CreateNotification")

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

// --- Trade Event Error Handling ---

func TestHandleTradeEvent_BuyerNotificationError(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.TradeEvent{
		TradeID:  "trade-1",
		PairID:   "BTC-USDT",
		BuyerID:  "buyer-1",
		SellerID: "seller-1",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	// First call fails for buyer, second succeeds for seller
	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "buyer-1"
	})).Return(false, fmt.Errorf("buyer notification failed")).Once()

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "seller-1"
	})).Return(true, nil).Once()

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trade notification errors")
}

func TestHandleTradeEvent_SellerNotificationError(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.TradeEvent{
		TradeID:  "trade-1",
		PairID:   "BTC-USDT",
		BuyerID:  "buyer-1",
		SellerID: "seller-1",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	// First call succeeds for buyer, second fails for seller
	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "buyer-1"
	})).Return(true, nil).Once()

	svc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.UserID == "seller-1"
	})).Return(false, fmt.Errorf("seller notification failed")).Once()

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trade notification errors")
}

func TestHandleTradeEvent_BothNotificationsError(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.TradeEvent{
		TradeID:  "trade-1",
		PairID:   "BTC-USDT",
		BuyerID:  "buyer-1",
		SellerID: "seller-1",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	// Both calls fail
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).
		Return(false, fmt.Errorf("notification failed")).Times(2)

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trade notification errors")
}

// --- Order Event Error Handling ---

func TestHandleOrderEvent_Filled_NotificationError(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.OrderEvent{
		OrderID:   "order-1",
		UserID:    "user-1",
		PairID:    "BTC-USDT",
		Side:      "buy",
		Price:     "50000",
		Quantity:  "1",
		FilledQty: "1",
		Status:    "filled",
		UpdatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    data,
	}

	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).
		Return(false, fmt.Errorf("notification error"))

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
}

// --- Wallet Event Error Handling ---

func TestHandleWalletEvent_Deposit_NotificationError(t *testing.T) {
	svc, handler := newTestEventHandler(t)

	event := pkgnats.WalletEvent{
		UserID:    "user-1",
		Asset:     "BTC",
		Amount:    "0.5",
		Action:    "deposit",
		TxID:      "tx-1",
		CreatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "wallets.user-1.deposit",
		Data:    data,
	}

	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).
		Return(false, fmt.Errorf("notification error"))

	err := handler.Handle(context.Background(), msg)
	assert.Error(t, err)
}
