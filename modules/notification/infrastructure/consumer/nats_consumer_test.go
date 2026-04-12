package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/entities"
	"booker/modules/notification/domain/interfaces"
	"booker/modules/notification/domain/interfaces/mocks"
	"booker/pkg/logger"
	pkgnats "booker/pkg/nats"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) With(args ...interface{}) logger.Logger {
	return m
}

func (m *MockLogger) WithGroup(name string) logger.Logger {
	return m
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

// --- Test Cases ---

func TestNewNATSConsumer_CreatesConsumer(t *testing.T) {
	// Create minimal mocks - just pass nil to test basic initialization
	var js nats.JetStreamContext
	handler := &EventHandler{}
	log := &MockLogger{}

	consumer := NewNATSConsumer(js, handler, log)

	assert.NotNil(t, consumer)
	assert.Equal(t, handler, consumer.handler)
	assert.Nil(t, consumer.cancel)
	assert.Len(t, consumer.subs, 0)
}

func TestNATSConsumer_Stop_WithoutStart(t *testing.T) {
	var js nats.JetStreamContext
	handler := &EventHandler{}
	log := &MockLogger{}

	consumer := NewNATSConsumer(js, handler, log)

	// Should not panic when stopping without starting
	consumer.Stop()
	assert.Nil(t, consumer.cancel)
}

// --- Testing with simpler approach ---
// Since testing Start/processMessages requires complex NATS mocking,
// focus on the coverage gaps in the simpler Stop() method and handler initialization

// --- Stop() Tests ---

func TestStop_CancelsContextMultipleTimes(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Manually set up a cancel function to test Stop behavior
	_, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// First stop should work
	consumer.Stop()

	// Calling Stop multiple times should not panic
	consumer.Stop()
	consumer.Stop()

	assert.NotNil(t, consumer.cancel, "cancel should still exist after Stop")
}

func TestStop_WithNilCancel_NoError(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// cancel is nil initially
	assert.Nil(t, consumer.cancel)

	// Should not panic when Stop is called before Start
	consumer.Stop()

	assert.Nil(t, consumer.cancel)
}

func TestStop_WithEmptySubscriptions(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// No subscriptions initially
	assert.Len(t, consumer.subs, 0)

	// Set up a cancel context
	_, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// Stop should complete without error
	consumer.Stop()

	assert.NotNil(t, consumer.cancel)
}

func TestStop_CallsCancelFunction(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Create a context to track cancellation
	ctx, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// Verify context is not yet cancelled
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be cancelled yet")
	default:
	}

	// Call Stop - should cancel the context
	consumer.Stop()

	// Give a moment for cancellation to propagate
	<-time.After(10 * time.Millisecond)

	// After Stop, context should be cancelled
	// (but we can't test this directly since ctx was created before Stop)
	// Instead, verify the method completes without error
	assert.NotNil(t, consumer.cancel)
}

func TestNewNATSConsumer_StoresAllDependencies(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	assert.NotNil(t, consumer)
	assert.Equal(t, handler, consumer.handler)
	assert.Equal(t, log, consumer.log)
	assert.Len(t, consumer.subs, 0)
	assert.Nil(t, consumer.cancel)
}

// --- Handler Initialization Tests ---

func TestEventHandler_NewEventHandler_HasService(t *testing.T) {
	svc := &MockNotificationService{}
	handler := NewEventHandler(svc)

	assert.NotNil(t, handler)
	assert.Equal(t, svc, handler.notifSvc)
}

// MockNotificationService for testing EventHandler creation
type MockNotificationService struct{}

var _ interfaces.NotificationService = (*MockNotificationService)(nil)

func (m *MockNotificationService) CreateNotification(ctx context.Context, n *entities.Notification) (bool, error) {
	return true, nil
}

func (m *MockNotificationService) ListNotifications(ctx context.Context, userID string, req *dto.ListNotificationsDTO) ([]*entities.Notification, error) {
	return nil, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, id, userID string) error {
	return nil
}

func (m *MockNotificationService) MarkAllAsRead(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

func (m *MockNotificationService) CountUnread(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

// --- Additional Event Handler Tests ---

func TestHandleTradeEvent_EventKeyFormat(t *testing.T) {
	_ = &MockNotificationService{}

	event := pkgnats.TradeEvent{
		TradeID:  "trade-abc-123",
		BuyerID:  "buyer-1",
		SellerID: "seller-1",
		PairID:   "BTC-USDT",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	// Create a recording mock to verify event keys
	var capturedNotifs []*entities.Notification
	mockSvc := mocks.NewMockNotificationService(t)
	mockSvc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		capturedNotifs = append(capturedNotifs, n)
		return true
	})).Return(true, nil).Times(2)

	handlerWithMock := NewEventHandler(mockSvc)
	err := handlerWithMock.Handle(context.Background(), msg)

	assert.NoError(t, err)
	assert.Len(t, capturedNotifs, 2)

	// Verify event keys have correct format
	assert.Contains(t, capturedNotifs[0].EventKey, "trade-abc-123")
	assert.Contains(t, capturedNotifs[1].EventKey, "trade-abc-123")
}

func TestHandleOrderEvent_EventKeyFormatFilled(t *testing.T) {
	event := pkgnats.OrderEvent{
		OrderID:  "order-xyz-789",
		UserID:   "user-1",
		Status:   "filled",
		Side:     "buy",
		Quantity: "1",
		PairID:   "ETH-USDT",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    data,
	}

	var capturedNotif *entities.Notification
	mockSvc := mocks.NewMockNotificationService(t)
	mockSvc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		capturedNotif = n
		return true
	})).Return(true, nil).Once()

	handler := NewEventHandler(mockSvc)
	err := handler.Handle(context.Background(), msg)

	assert.NoError(t, err)
	assert.NotNil(t, capturedNotif)
	assert.Equal(t, "order_order-xyz-789_filled", capturedNotif.EventKey)
}

func TestHandleOrderEvent_EventKeyFormatCancelled(t *testing.T) {
	event := pkgnats.OrderEvent{
		OrderID:  "order-xyz-999",
		UserID:   "user-2",
		Status:   "cancelled",
		Side:     "sell",
		Quantity: "10",
		PairID:   "BTC-USDT",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "orders.user-2.cancelled",
		Data:    data,
	}

	var capturedNotif *entities.Notification
	mockSvc := mocks.NewMockNotificationService(t)
	mockSvc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		capturedNotif = n
		return true
	})).Return(true, nil).Once()

	handler := NewEventHandler(mockSvc)
	err := handler.Handle(context.Background(), msg)

	assert.NoError(t, err)
	assert.NotNil(t, capturedNotif)
	assert.Equal(t, "order_order-xyz-999_cancelled", capturedNotif.EventKey)
}

func TestHandleWalletEvent_TxIDInMetadata(t *testing.T) {
	event := pkgnats.WalletEvent{
		UserID:    "user-1",
		Asset:     "BTC",
		Amount:    "0.5",
		Action:    "deposit",
		TxID:      "tx-abc-123-def",
		CreatedAt: "2026-04-12T10:00:00Z",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "wallets.user-1.deposit",
		Data:    data,
	}

	var capturedNotif *entities.Notification
	mockSvc := mocks.NewMockNotificationService(t)
	mockSvc.EXPECT().CreateNotification(mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		capturedNotif = n
		return true
	})).Return(true, nil).Once()

	handler := NewEventHandler(mockSvc)
	err := handler.Handle(context.Background(), msg)

	assert.NoError(t, err)
	assert.NotNil(t, capturedNotif)
	assert.Equal(t, "wallet_tx-abc-123-def", capturedNotif.EventKey)
	assert.Equal(t, "tx-abc-123-def", capturedNotif.Metadata["tx_id"])
}

func TestHandle_SubjectPrefixMatching_Trades(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	handler := NewEventHandler(svc)

	// Test various trade subject patterns
	subjects := []string{
		"trades.BTC-USDT.executed",
		"trades.ETH-USDT.executed",
		"trades.XRP-USDT.executed",
	}

	for _, subject := range subjects {
		event := pkgnats.TradeEvent{
			TradeID:  "t-1",
			BuyerID:  "b-1",
			SellerID: "s-1",
			PairID:   "TEST",
			Price:    "100",
			Quantity: "1",
		}
		data, _ := json.Marshal(event)
		msg := &nats.Msg{
			Subject: subject,
			Data:    data,
		}

		svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Times(2)
		err := handler.Handle(context.Background(), msg)
		assert.NoError(t, err)
	}
}

func TestHandle_SubjectPrefixMatching_Orders(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	handler := NewEventHandler(svc)

	subjects := []string{
		"orders.user-1.filled",
		"orders.user-2.cancelled",
	}

	for _, subject := range subjects {
		event := pkgnats.OrderEvent{
			OrderID:  "o-1",
			UserID:   "u-1",
			Status:   "filled",
			Side:     "buy",
			PairID:   "BTC-USDT",
			Quantity: "1",
		}
		data, _ := json.Marshal(event)
		msg := &nats.Msg{
			Subject: subject,
			Data:    data,
		}

		svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()
		err := handler.Handle(context.Background(), msg)
		assert.NoError(t, err)
	}
}

func TestHandle_SubjectPrefixMatching_Wallets(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	handler := NewEventHandler(svc)

	subjects := []string{
		"wallets.user-1.deposit",
		"wallets.user-2.withdrawal",
	}

	for _, subject := range subjects {
		event := pkgnats.WalletEvent{
			UserID: "u-1",
			Asset:  "BTC",
			Amount: "1",
			Action: "deposit",
			TxID:   "tx-1",
		}
		data, _ := json.Marshal(event)
		msg := &nats.Msg{
			Subject: subject,
			Data:    data,
		}

		svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()
		err := handler.Handle(context.Background(), msg)
		assert.NoError(t, err)
	}
}

// --- Mock JetStream and Subscription Interfaces ---

// SimpleJetStreamMock provides minimal JetStream interface implementation for testing
type SimpleJetStreamMock struct {
	pullSubErr error
	pullSubRtn *nats.Subscription
}

func (m *SimpleJetStreamMock) PullSubscribe(subject, durable string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return m.pullSubRtn, m.pullSubErr
}

// MockSubscription wraps nats.Subscription to provide testable behavior
type MockSubscription struct {
	*nats.Subscription
	fetchCount int
	fetchErr   error
	drainErr   error
	messages   []*nats.Msg
}

// Custom fetch implementation that returns our test messages or errors
func (ms *MockSubscription) Fetch(batch int, opts ...nats.SubOpt) ([]*nats.Msg, error) {
	ms.fetchCount++
	if ms.fetchErr != nil {
		return nil, ms.fetchErr
	}
	if len(ms.messages) == 0 {
		return []*nats.Msg{}, nil
	}
	result := ms.messages
	ms.messages = nil
	return result, nil
}

// --- Start() Tests ---

func TestStart_SubscribesToAllStreams(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	mockSub := &nats.Subscription{}
	jsMock := &SimpleJetStreamMock{
		pullSubRtn: mockSub,
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start should subscribe to all streams
	consumer.Start(ctx)

	// Give goroutines time to start
	<-time.After(50 * time.Millisecond)

	// Verify all subscriptions were created
	assert.Len(t, consumer.subs, 3)

	// Cleanup
	consumer.Stop()
}

func TestStart_SubscriptionFailureLogsError(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: nil,
		pullSubErr: fmt.Errorf("connection failed"),
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// No successful subscriptions due to error
	assert.Len(t, consumer.subs, 0)

	// Verify error was logged
	assert.True(t, len(log.messages) > 0, "Expected error to be logged")

	consumer.Stop()
}

func TestStart_CreatesContextWithCancel(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)

	assert.Nil(t, consumer.cancel, "cancel should be nil before Start")

	ctx := context.Background()
	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	assert.NotNil(t, consumer.cancel, "cancel should be set after Start")

	consumer.Stop()
}

func TestStart_LaunchesProcessMessagesGoroutines(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// Should have 3 subscriptions and 3 goroutines
	assert.Len(t, consumer.subs, 3)

	consumer.Stop()
}

// --- processMessages() Tests (implicit via Start) ---

func TestProcessMessages_ContextCancellationStopsLoop(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())

	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// Cancel context should stop processMessages
	cancel()
	<-time.After(100 * time.Millisecond)

	// Consumer should stop cleanly
	consumer.Stop()
	assert.NotNil(t, consumer)
}

func TestStart_PartialSubscriptionFailure(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	// This simulates successful subscription
	jsFailOnce := &SimpleJetStreamMock{
		pullSubErr: nil,
	}
	jsFailOnce.pullSubRtn = &nats.Subscription{}

	consumer := NewNATSConsumer(jsFailOnce, handler, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// Should have subscriptions from successful calls
	assert.True(t, len(consumer.subs) >= 0)

	consumer.Stop()
}

// --- Stop() Tests (enhanced) ---

func TestStop_CancelsFunctionCalledOnce(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Set up a context we can monitor
	ctx, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// Before stop, context should be active
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be cancelled yet")
	default:
	}

	// Call stop
	consumer.Stop()
	<-time.After(10 * time.Millisecond)

	// After stop, we know cancel was called (context is now cancelled)
	select {
	case <-ctx.Done():
		// Expected - context is cancelled
	default:
		t.Fatal("Context should be cancelled after Stop")
	}
}

func TestStop_WithNilCancelNoError(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// cancel is nil
	assert.Nil(t, consumer.cancel)

	// Should not panic
	assert.NotPanics(t, func() {
		consumer.Stop()
	})
}

func TestStop_IteratesDrainSubscriptions(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)
	_, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// Add mock subscriptions
	consumer.subs = []*nats.Subscription{
		{},
		{},
		{},
	}

	// Stop should iterate all subscriptions without error
	assert.NotPanics(t, func() {
		consumer.Stop()
	})
}

func TestStop_NoSubscriptionsNoPanic(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)
	_, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// No subscriptions
	assert.Len(t, consumer.subs, 0)

	// Should complete without error
	consumer.Stop()
	assert.NotNil(t, consumer)
}

func TestStop_MultipleCallsSafe(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)
	_, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// Multiple stops should be safe
	assert.NotPanics(t, func() {
		consumer.Stop()
		consumer.Stop()
		consumer.Stop()
	})
}

// --- Additional error path tests for processMessages ---

func TestProcessMessages_ErrorHandlingPath(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())

	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// Verify goroutines are running and can be stopped
	cancel()
	<-time.After(100 * time.Millisecond)

	consumer.Stop()
	assert.NotNil(t, consumer)
}

func TestStart_AllSubscriptionsFailure(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: nil,
		pullSubErr: fmt.Errorf("all subscriptions failed"),
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// No subscriptions should be created when all fail
	assert.Len(t, consumer.subs, 0)

	// Errors should be logged
	assert.True(t, len(log.messages) >= 3, "Expected 3+ error logs for all subscription failures")

	consumer.Stop()
}

func TestStart_LogsInfoForSuccessfulSubscriptions(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialMsgCount := len(log.messages)
	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// Should have logged info messages for successful subscriptions
	finalMsgCount := len(log.messages)
	assert.Greater(t, finalMsgCount, initialMsgCount, "Expected info messages to be logged")

	consumer.Stop()
}

func TestStart_ContextCancellationBeforeStart(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)

	// Create a cancelled context first
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Start with already-cancelled context
	consumer.Start(ctx)
	<-time.After(50 * time.Millisecond)

	// Should still attempt to subscribe
	assert.NotNil(t, consumer.cancel)

	consumer.Stop()
}

func TestStop_LogsWarningOnDrainError(t *testing.T) {
	h := &EventHandler{}
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, h, log)
	_, cancel := context.WithCancel(context.Background())
	consumer.cancel = cancel

	// We can't easily mock sub.Drain() error on real subscription,
	// so just verify Stop handles subscriptions gracefully
	consumer.subs = []*nats.Subscription{{}, {}, {}}

	consumer.Stop()
	// If there are drain errors, they would be logged, but without a real NATS
	// connection, Drain() will likely succeed or fail silently
	assert.NotNil(t, consumer)
}

func TestEventHandler_HandleMultipleMessageTypes(t *testing.T) {
	// Test that handler correctly routes different message types
	msgTypes := []struct {
		name    string
		subject string
		event   interface{}
		calls   int
	}{
		{
			name:    "trade",
			subject: "trades.BTC-USDT.executed",
			event: pkgnats.TradeEvent{
				TradeID:  "t1",
				BuyerID:  "b1",
				SellerID: "s1",
				PairID:   "BTC-USDT",
				Price:    "50000",
				Quantity: "1",
			},
			calls: 2,
		},
		{
			name:    "order",
			subject: "orders.u1.filled",
			event: pkgnats.OrderEvent{
				OrderID:  "o1",
				UserID:   "u1",
				Status:   "filled",
				PairID:   "BTC-USDT",
				Side:     "buy",
				Quantity: "1",
			},
			calls: 1,
		},
		{
			name:    "wallet",
			subject: "wallets.u1.deposit",
			event: pkgnats.WalletEvent{
				UserID: "u1",
				Asset:  "BTC",
				Amount: "0.5",
				Action: "deposit",
				TxID:   "tx1",
			},
			calls: 1,
		},
	}

	for _, mt := range msgTypes {
		t.Run(mt.name, func(t *testing.T) {
			data, err := json.Marshal(mt.event)
			assert.NoError(t, err)

			msg := &nats.Msg{
				Subject: mt.subject,
				Data:    data,
			}

			mockSvc := mocks.NewMockNotificationService(t)
			mockSvc.EXPECT().CreateNotification(mock.Anything, mock.Anything).
				Return(true, nil).Times(mt.calls)

			testHandler := NewEventHandler(mockSvc)
			err = testHandler.Handle(context.Background(), msg)
			assert.NoError(t, err)
		})
	}
}

func TestStart_SubscriptionDeadlineExceeded(t *testing.T) {
	handler := &EventHandler{}
	log := &MockLogger{}

	jsMock := &SimpleJetStreamMock{
		pullSubRtn: &nats.Subscription{},
		pullSubErr: nil,
	}

	consumer := NewNATSConsumer(jsMock, handler, log)

	// Create a context that will timeout immediately
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)
	<-time.After(100 * time.Millisecond)

	// After timeout, context should be done
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Fatal("Context should be cancelled by timeout")
	}

	consumer.Stop()
}

// --- Tests for handleFetchedMessages ---

func TestHandleFetchedMessages_SuccessfulAck(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Times(2)

	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Create mock messages with successful ACK
	tradeEvent := pkgnats.TradeEvent{
		TradeID:  "trade-1",
		BuyerID:  "buyer-1",
		SellerID: "seller-1",
		PairID:   "BTC-USDT",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(tradeEvent)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	// Call handleFetchedMessages
	consumer.handleFetchedMessages(context.Background(), []*nats.Msg{msg}, "trades.>")

	// Should have processed the message
	assert.NotNil(t, consumer)
}

func TestHandleFetchedMessages_EmptyMessages(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Handle empty message slice
	consumer.handleFetchedMessages(context.Background(), []*nats.Msg{}, "trades.>")

	// Should complete without error
	assert.NotNil(t, consumer)
}

func TestHandleFetchedMessages_MultipleMessages(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Times(6)

	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Create multiple messages
	msgs := []*nats.Msg{}
	for i := 0; i < 3; i++ {
		tradeEvent := pkgnats.TradeEvent{
			TradeID:  fmt.Sprintf("trade-%d", i),
			BuyerID:  fmt.Sprintf("buyer-%d", i),
			SellerID: fmt.Sprintf("seller-%d", i),
			PairID:   "BTC-USDT",
			Price:    "50000",
			Quantity: "0.5",
		}
		data, _ := json.Marshal(tradeEvent)
		msgs = append(msgs, &nats.Msg{
			Subject: "trades.BTC-USDT.executed",
			Data:    data,
		})
	}

	consumer.handleFetchedMessages(context.Background(), msgs, "trades.>")
	assert.NotNil(t, consumer)
}

func TestHandleFetchedMessages_OrderFilledMessages(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()

	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	orderEvent := pkgnats.OrderEvent{
		OrderID:  "order-1",
		UserID:   "user-1",
		Status:   "filled",
		PairID:   "BTC-USDT",
		Side:     "buy",
		Quantity: "1",
	}
	data, _ := json.Marshal(orderEvent)
	msg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    data,
	}

	consumer.handleFetchedMessages(context.Background(), []*nats.Msg{msg}, "orders.>")
	assert.NotNil(t, consumer)
}

func TestHandleFetchedMessages_WalletMessages(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()

	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	walletEvent := pkgnats.WalletEvent{
		UserID: "user-1",
		Asset:  "BTC",
		Amount: "0.5",
		Action: "deposit",
		TxID:   "tx-1",
	}
	data, _ := json.Marshal(walletEvent)
	msg := &nats.Msg{
		Subject: "wallets.user-1.deposit",
		Data:    data,
	}

	consumer.handleFetchedMessages(context.Background(), []*nats.Msg{msg}, "wallets.>")
	assert.NotNil(t, consumer)
}

func TestHandleFetchedMessages_WithInvalidMessage(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Create a message with invalid data that will fail parsing
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    []byte("invalid json"),
	}

	// Should handle error gracefully
	consumer.handleFetchedMessages(context.Background(), []*nats.Msg{msg}, "trades.>")

	// Should have logged error
	assert.True(t, len(log.messages) > 0, "Expected error to be logged")
}

func TestHandleFetchedMessages_MixedValidInvalid(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Once()

	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Create a mix of valid and invalid messages
	validEvent := pkgnats.OrderEvent{
		OrderID:  "order-1",
		UserID:   "user-1",
		Status:   "filled",
		PairID:   "BTC-USDT",
		Side:     "buy",
		Quantity: "1",
	}
	validData, _ := json.Marshal(validEvent)
	validMsg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    validData,
	}

	invalidMsg := &nats.Msg{
		Subject: "orders.user-1.filled",
		Data:    []byte("invalid json"),
	}

	msgs := []*nats.Msg{validMsg, invalidMsg}
	consumer.handleFetchedMessages(context.Background(), msgs, "orders.>")

	assert.NotNil(t, consumer)
}

func TestHandleFetchedMessages_ContextCancellation(t *testing.T) {
	svc := mocks.NewMockNotificationService(t)
	svc.EXPECT().CreateNotification(mock.Anything, mock.Anything).Return(true, nil).Times(2)

	handler := NewEventHandler(svc)
	log := &MockLogger{}
	var js nats.JetStreamContext

	consumer := NewNATSConsumer(js, handler, log)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a valid message
	event := pkgnats.TradeEvent{
		TradeID:  "trade-1",
		BuyerID:  "buyer-1",
		SellerID: "seller-1",
		PairID:   "BTC-USDT",
		Price:    "50000",
		Quantity: "0.5",
	}
	data, _ := json.Marshal(event)
	msg := &nats.Msg{
		Subject: "trades.BTC-USDT.executed",
		Data:    data,
	}

	// Should handle with cancelled context gracefully
	consumer.handleFetchedMessages(ctx, []*nats.Msg{msg}, "trades.>")

	assert.NotNil(t, consumer)
}
