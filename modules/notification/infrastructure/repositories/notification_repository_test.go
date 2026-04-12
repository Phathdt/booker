package repositories

import (
	"context"
	"testing"
	"time"

	"booker/modules/notification/domain"
	"booker/modules/notification/domain/entities"
	tc "booker/test/testcontainers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	repo := NewNotificationRepository(containers.Database)
	ctx := context.Background()

	userID := uuid.New().String()

	t.Run("Create notification", func(t *testing.T) {
		n := &entities.Notification{
			UserID:   userID,
			EventKey: "trade_abc_" + userID,
			Type:     entities.TypeTradeExecuted,
			Title:    "Trade Executed",
			Body:     "Your buy order was filled at 50000 USDT",
			Metadata: map[string]string{"pair": "BTC_USDT", "price": "50000"},
		}
		ok, err := repo.Create(ctx, n)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Create duplicate event_key is deduplicated", func(t *testing.T) {
		n := &entities.Notification{
			UserID:   userID,
			EventKey: "trade_abc_" + userID,
			Type:     entities.TypeTradeExecuted,
			Title:    "Trade Executed Duplicate",
			Body:     "Should be ignored",
			Metadata: map[string]string{},
		}
		ok, err := repo.Create(ctx, n)
		require.NoError(t, err)
		assert.False(t, ok) // ON CONFLICT DO NOTHING → 0 rows affected
	})

	t.Run("ListByUser returns notifications", func(t *testing.T) {
		// Create a second notification
		_, err := repo.Create(ctx, &entities.Notification{
			UserID:   userID,
			EventKey: "order_xyz_filled",
			Type:     entities.TypeOrderFilled,
			Title:    "Order Filled",
			Body:     "Your order was completely filled",
			Metadata: map[string]string{"order_id": "xyz"},
		})
		require.NoError(t, err)

		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, userID, cursor, 10, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notifications), 2)

		// Verify ordering: newest first
		for i := 1; i < len(notifications); i++ {
			assert.True(t, notifications[i-1].CreatedAt.After(notifications[i].CreatedAt) ||
				notifications[i-1].CreatedAt.Equal(notifications[i].CreatedAt))
		}
	})

	t.Run("ListByUser with onlyUnread", func(t *testing.T) {
		cursor := time.Now().Add(time.Minute)
		unread, err := repo.ListByUser(ctx, userID, cursor, 10, true)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(unread), 2)
		for _, n := range unread {
			assert.False(t, n.IsRead)
		}
	})

	t.Run("CountUnread", func(t *testing.T) {
		count, err := repo.CountUnread(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, userID, cursor, 1, false)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)

		err = repo.MarkAsRead(ctx, notifications[0].ID, userID)
		require.NoError(t, err)

		// Verify unread count decreased
		count, err := repo.CountUnread(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("MarkAsRead not found", func(t *testing.T) {
		err := repo.MarkAsRead(ctx, uuid.New().String(), userID)
		assert.Equal(t, domain.ErrNotificationNotFound, err)
	})

	t.Run("MarkAsRead wrong user", func(t *testing.T) {
		// Create a self-contained notification for this test
		_, err := repo.Create(ctx, &entities.Notification{
			UserID:   userID,
			EventKey: "wrong_user_test_" + uuid.New().String(),
			Type:     entities.TypeDepositConfirmed,
			Title:    "Test Notification",
			Body:     "For wrong user test",
			Metadata: map[string]string{},
		})
		require.NoError(t, err)

		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, userID, cursor, 1, true)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)

		err = repo.MarkAsRead(ctx, notifications[0].ID, uuid.New().String())
		assert.Equal(t, domain.ErrNotificationNotFound, err)
	})

	t.Run("MarkAllAsRead", func(t *testing.T) {
		affected, err := repo.MarkAllAsRead(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, affected, int64(1))

		count, err := repo.CountUnread(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("MarkAllAsRead no unread", func(t *testing.T) {
		affected, err := repo.MarkAllAsRead(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected)
	})

	t.Run("ListByUser empty for unknown user", func(t *testing.T) {
		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, uuid.New().String(), cursor, 10, false)
		require.NoError(t, err)
		assert.Empty(t, notifications)
	})

	t.Run("CountUnread for unknown user", func(t *testing.T) {
		count, err := repo.CountUnread(ctx, uuid.New().String())
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Create with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Create(cancelCtx, &entities.Notification{
			UserID:   userID,
			EventKey: "cancelled_ctx_test",
			Type:     entities.TypeDepositConfirmed,
			Title:    "Test",
			Body:     "Test",
			Metadata: map[string]string{},
		})
		assert.Error(t, err)
	})
}
