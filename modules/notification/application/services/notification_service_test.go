package services

import (
	"context"
	"testing"
	"time"

	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/entities"
	"booker/modules/notification/domain/interfaces"
	"booker/modules/notification/domain/interfaces/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestNotificationService(t *testing.T) (*mocks.MockNotificationRepository, *mocks.MockBroadcaster, interfaces.NotificationService) {
	repo := mocks.NewMockNotificationRepository(t)
	broadcaster := mocks.NewMockBroadcaster(t)
	svc := NewNotificationService(repo, broadcaster)
	return repo, broadcaster, svc
}

// --- CreateNotification Tests ---

func TestCreateNotification_Success(t *testing.T) {
	repo, broadcaster, svc := newTestNotificationService(t)

	notif := &entities.Notification{
		UserID:   "user-1",
		EventKey: "trade_t1_user-1",
		Type:     entities.TypeTradeExecuted,
		Title:    "Trade Executed",
		Body:     "You bought 0.5 BTC at 50000",
		Metadata: map[string]string{"trade_id": "t1"},
	}

	repo.EXPECT().Create(mock.Anything, notif).Return(true, nil).Once()
	broadcaster.EXPECT().SendToUser("user-1", notif).Once()

	inserted, err := svc.CreateNotification(context.Background(), notif)
	assert.NoError(t, err)
	assert.True(t, inserted)
}

func TestCreateNotification_Duplicate(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	notif := &entities.Notification{
		UserID:   "user-1",
		EventKey: "trade_t1_user-1",
		Type:     entities.TypeTradeExecuted,
	}

	// Duplicate: repo returns false (no row inserted), broadcaster should NOT be called
	repo.EXPECT().Create(mock.Anything, notif).Return(false, nil).Once()

	inserted, err := svc.CreateNotification(context.Background(), notif)
	assert.NoError(t, err)
	assert.False(t, inserted)
}

func TestCreateNotification_RepositoryError(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	notif := &entities.Notification{
		UserID: "user-1",
		Type:   entities.TypeTradeExecuted,
	}

	repo.EXPECT().Create(mock.Anything, notif).Return(false, assert.AnError).Once()

	_, err := svc.CreateNotification(context.Background(), notif)
	assert.Error(t, err)
}

func TestCreateNotification_BroadcasterNil(t *testing.T) {
	repo := mocks.NewMockNotificationRepository(t)
	svc := NewNotificationService(repo, nil)

	notif := &entities.Notification{
		UserID:   "user-1",
		EventKey: "order_o1_filled",
		Type:     entities.TypeOrderFilled,
		Title:    "Order Filled",
		Body:     "Your order has been filled",
	}

	repo.EXPECT().Create(mock.Anything, notif).Return(true, nil).Once()

	inserted, err := svc.CreateNotification(context.Background(), notif)
	assert.NoError(t, err)
	assert.True(t, inserted)
}

// --- ListNotifications Tests ---

func TestListNotifications_DefaultLimit(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	notifs := []*entities.Notification{
		{
			ID:        "notif-1",
			UserID:    "user-1",
			Type:      entities.TypeTradeExecuted,
			Title:     "Trade 1",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "notif-2",
			UserID:    "user-1",
			Type:      entities.TypeOrderFilled,
			Title:     "Trade 2",
			CreatedAt: time.Now(),
		},
	}

	req := &dto.ListNotificationsDTO{}
	repo.EXPECT().ListByUser(mock.Anything, "user-1", mock.MatchedBy(func(cursor time.Time) bool {
		// Cursor should be close to now
		return time.Since(cursor) < 1*time.Second
	}), int32(20), false).Return(notifs, nil).Once()

	result, err := svc.ListNotifications(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "notif-1", result[0].ID)
}

func TestListNotifications_CustomLimit(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	notifs := []*entities.Notification{
		{ID: "notif-1", UserID: "user-1"},
	}

	req := &dto.ListNotificationsDTO{Limit: 10}
	repo.EXPECT().ListByUser(mock.Anything, "user-1", mock.MatchedBy(func(cursor time.Time) bool {
		return time.Since(cursor) < 1*time.Second
	}), int32(10), false).Return(notifs, nil).Once()

	result, err := svc.ListNotifications(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestListNotifications_CustomCursor(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	cursorTime := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	cursorStr := cursorTime.Format(time.RFC3339)

	notifs := []*entities.Notification{}
	req := &dto.ListNotificationsDTO{
		Cursor: cursorStr,
		Limit:  15,
	}
	repo.EXPECT().ListByUser(mock.Anything, "user-1", mock.MatchedBy(func(cursor time.Time) bool {
		return cursor.Equal(cursorTime)
	}), int32(15), false).Return(notifs, nil).Once()

	result, err := svc.ListNotifications(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestListNotifications_InvalidCursor(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	notifs := []*entities.Notification{}
	req := &dto.ListNotificationsDTO{Cursor: "invalid-date"}

	// Invalid cursor should be ignored and default to now
	repo.EXPECT().ListByUser(mock.Anything, "user-1", mock.MatchedBy(func(cursor time.Time) bool {
		return time.Since(cursor) < 1*time.Second
	}), int32(20), false).Return(notifs, nil).Once()

	result, err := svc.ListNotifications(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestListNotifications_OnlyUnread(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	unreadNotifs := []*entities.Notification{
		{ID: "notif-1", UserID: "user-1", IsRead: false},
	}

	req := &dto.ListNotificationsDTO{OnlyUnread: true}
	repo.EXPECT().ListByUser(mock.Anything, "user-1", mock.MatchedBy(func(cursor time.Time) bool {
		return time.Since(cursor) < 1*time.Second
	}), int32(20), true).Return(unreadNotifs, nil).Once()

	result, err := svc.ListNotifications(context.Background(), "user-1", req)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.False(t, result[0].IsRead)
}

// --- MarkAsRead Tests ---

func TestMarkAsRead_Success(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	repo.EXPECT().MarkAsRead(mock.Anything, "notif-1", "user-1").Return(nil).Once()

	err := svc.MarkAsRead(context.Background(), "notif-1", "user-1")
	assert.NoError(t, err)
}

func TestMarkAsRead_RepositoryError(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	repo.EXPECT().MarkAsRead(mock.Anything, "notif-1", "user-1").Return(assert.AnError).Once()

	err := svc.MarkAsRead(context.Background(), "notif-1", "user-1")
	assert.Error(t, err)
}

// --- MarkAllAsRead Tests ---

func TestMarkAllAsRead_Success(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	repo.EXPECT().MarkAllAsRead(mock.Anything, "user-1").Return(int64(5), nil).Once()

	count, err := svc.MarkAllAsRead(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestMarkAllAsRead_Zero(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	repo.EXPECT().MarkAllAsRead(mock.Anything, "user-1").Return(int64(0), nil).Once()

	count, err := svc.MarkAllAsRead(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// --- CountUnread Tests ---

func TestCountUnread_Success(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	repo.EXPECT().CountUnread(mock.Anything, "user-1").Return(int64(3), nil).Once()

	count, err := svc.CountUnread(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestCountUnread_Zero(t *testing.T) {
	repo, _, svc := newTestNotificationService(t)

	repo.EXPECT().CountUnread(mock.Anything, "user-1").Return(int64(0), nil).Once()

	count, err := svc.CountUnread(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
