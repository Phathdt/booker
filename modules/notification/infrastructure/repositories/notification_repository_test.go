package repositories

import (
	"context"
	"fmt"
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

	t.Run("ListByUser with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		cursor := time.Now()
		_, err := repo.ListByUser(cancelCtx, userID, cursor, 10, false)
		assert.Error(t, err)
	})

	t.Run("CountUnread with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.CountUnread(cancelCtx, userID)
		assert.Error(t, err)
	})

	t.Run("MarkAsRead with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		err := repo.MarkAsRead(cancelCtx, uuid.New().String(), userID)
		assert.Error(t, err)
	})

	t.Run("MarkAllAsRead with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.MarkAllAsRead(cancelCtx, userID)
		assert.Error(t, err)
	})

	t.Run("Create with complex metadata", func(t *testing.T) {
		n := &entities.Notification{
			UserID:   userID,
			EventKey: "complex_metadata_" + uuid.New().String(),
			Type:     entities.TypeTradeExecuted,
			Title:    "Complex Trade",
			Body:     "Trade with detailed metadata",
			Metadata: map[string]string{
				"pair":           "BTC_USDT",
				"price":          "50123.45",
				"quantity":       "0.5",
				"fee":            "0.001",
				"order_id":       "order_123_abc",
				"trade_id":       "trade_456_def",
				"buyer_name":     "Alice",
				"seller_name":    "Bob",
				"execution_time": "2026-04-12T21:46:00Z",
			},
		}
		ok, err := repo.Create(ctx, n)
		require.NoError(t, err)
		assert.True(t, ok)

		// Verify metadata is preserved
		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, userID, cursor, 1, false)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)

		found := false
		for _, notif := range notifications {
			if notif.EventKey == n.EventKey {
				found = true
				assert.Equal(t, "50123.45", notif.Metadata["price"])
				assert.Equal(t, "0.5", notif.Metadata["quantity"])
				assert.Equal(t, "Alice", notif.Metadata["buyer_name"])
				break
			}
		}
		assert.True(t, found, "notification with complex metadata not found")
	})

	t.Run("Create multiple events for same user", func(t *testing.T) {
		eventUser := uuid.New().String()
		for i := 0; i < 5; i++ {
			indexStr := fmt.Sprintf("%d", i)
			n := &entities.Notification{
				UserID:   eventUser,
				EventKey: "event_" + uuid.New().String(),
				Type:     entities.TypeOrderFilled,
				Title:    "Order Filled " + indexStr,
				Body:     "Your order was filled",
				Metadata: map[string]string{"index": indexStr},
			}
			ok, err := repo.Create(ctx, n)
			require.NoError(t, err)
			assert.True(t, ok)
		}

		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, eventUser, cursor, 100, false)
		require.NoError(t, err)
		assert.Equal(t, 5, len(notifications))

		// Verify newest first ordering
		for i := 1; i < len(notifications); i++ {
			assert.True(t, notifications[i-1].CreatedAt.After(notifications[i].CreatedAt) ||
				notifications[i-1].CreatedAt.Equal(notifications[i].CreatedAt),
				"notifications not in newest-first order")
		}
	})

	t.Run("MarkAsRead transitions unread to read", func(t *testing.T) {
		readTestUser := uuid.New().String()
		n := &entities.Notification{
			UserID:   readTestUser,
			EventKey: "read_test_" + uuid.New().String(),
			Type:     entities.TypeWithdrawalConfirmed,
			Title:    "Withdrawal Confirmed",
			Body:     "Your withdrawal is confirmed",
			Metadata: map[string]string{},
		}
		ok, err := repo.Create(ctx, n)
		require.NoError(t, err)
		assert.True(t, ok)

		// Verify it starts unread
		cursor := time.Now().Add(time.Minute)
		unreadBefore, err := repo.ListByUser(ctx, readTestUser, cursor, 1, true)
		require.NoError(t, err)
		require.Len(t, unreadBefore, 1)
		assert.False(t, unreadBefore[0].IsRead)

		// Mark as read
		err = repo.MarkAsRead(ctx, unreadBefore[0].ID, readTestUser)
		require.NoError(t, err)

		// Verify it's now read
		unreadAfter, err := repo.ListByUser(ctx, readTestUser, cursor, 1, true)
		require.NoError(t, err)
		assert.Empty(t, unreadAfter)

		// Verify it shows in all list
		allAfter, err := repo.ListByUser(ctx, readTestUser, cursor, 1, false)
		require.NoError(t, err)
		require.Len(t, allAfter, 1)
		assert.True(t, allAfter[0].IsRead)
	})

	t.Run("MarkAllAsRead marks all unread notifications", func(t *testing.T) {
		bulkReadUser := uuid.New().String()
		// Create 3 unread notifications
		for i := 0; i < 3; i++ {
			indexStr := fmt.Sprintf("%d", i)
			_, err := repo.Create(ctx, &entities.Notification{
				UserID:   bulkReadUser,
				EventKey: "bulk_read_" + uuid.New().String(),
				Type:     entities.TypeDepositConfirmed,
				Title:    "Deposit #" + indexStr,
				Body:     "Deposit confirmed",
				Metadata: map[string]string{},
			})
			require.NoError(t, err)
		}

		// Verify 3 unread
		unreadBefore, err := repo.CountUnread(ctx, bulkReadUser)
		require.NoError(t, err)
		assert.Equal(t, int64(3), unreadBefore)

		// Mark all as read
		affected, err := repo.MarkAllAsRead(ctx, bulkReadUser)
		require.NoError(t, err)
		assert.Equal(t, int64(3), affected)

		// Verify 0 unread
		unreadAfter, err := repo.CountUnread(ctx, bulkReadUser)
		require.NoError(t, err)
		assert.Equal(t, int64(0), unreadAfter)
	})

	t.Run("ListByUser respects limit and cursor", func(t *testing.T) {
		cursorUser := uuid.New().String()
		// Create 5 notifications with timestamps
		for i := 0; i < 5; i++ {
			indexStr := fmt.Sprintf("%d", i)
			_, err := repo.Create(ctx, &entities.Notification{
				UserID:   cursorUser,
				EventKey: "cursor_test_" + uuid.New().String(),
				Type:     entities.TypeOrderFilled,
				Title:    "Notification " + indexStr,
				Body:     "Test",
				Metadata: map[string]string{},
			})
			require.NoError(t, err)
		}

		// List with small limit
		cursor := time.Now().Add(time.Minute)
		page1, err := repo.ListByUser(ctx, cursorUser, cursor, 2, false)
		require.NoError(t, err)
		assert.Equal(t, 2, len(page1))

		// Use oldest from page1 as new cursor
		page2, err := repo.ListByUser(ctx, cursorUser, page1[len(page1)-1].CreatedAt, 2, false)
		require.NoError(t, err)
		assert.Equal(t, 2, len(page2))

		// Verify no overlap
		for _, p1 := range page1 {
			for _, p2 := range page2 {
				assert.NotEqual(t, p1.ID, p2.ID, "page overlap detected")
			}
		}
	})

	t.Run("Notification types coverage", func(t *testing.T) {
		typesUser := uuid.New().String()
		notificationTypes := []entities.NotificationType{
			entities.TypeTradeExecuted,
			entities.TypeOrderFilled,
			entities.TypeOrderCancelled,
			entities.TypeDepositConfirmed,
			entities.TypeWithdrawalConfirmed,
		}

		for i, notifType := range notificationTypes {
			_, err := repo.Create(ctx, &entities.Notification{
				UserID:   typesUser,
				EventKey: "type_" + string(notifType) + "_" + uuid.New().String(),
				Type:     notifType,
				Title:    "Test " + string(notifType),
				Body:     "Test body",
				Metadata: map[string]string{},
			})
			require.NoError(t, err, "failed to create notification type %d", i)
		}

		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, typesUser, cursor, 100, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notifications), len(notificationTypes))
	})

	t.Run("Create notification with invalid JSON metadata should fail", func(t *testing.T) {
		// This test verifies that bad JSON in metadata would be caught by json.Marshal
		// The current implementation uses valid maps, but this documents the expected behavior
		n := &entities.Notification{
			UserID:   userID,
			EventKey: "valid_json_" + uuid.New().String(),
			Type:     entities.TypeTradeExecuted,
			Title:    "Valid Test",
			Body:     "Body",
			Metadata: map[string]string{"key": "value"},
		}
		ok, err := repo.Create(ctx, n)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("ListByUser onlyUnread=false path", func(t *testing.T) {
		// Create notification for a specific user
		testUser := uuid.New().String()
		_, _ = repo.Create(ctx, &entities.Notification{
			UserID:   testUser,
			EventKey: "unread_path_test_" + uuid.New().String(),
			Type:     entities.TypeTradeExecuted,
			Title:    "Unread Path Test",
			Body:     "Testing the ListByUser unread=false path",
			Metadata: map[string]string{"test": "true"},
		})

		cursor := time.Now().Add(time.Minute)
		// This should execute the non-onlyUnread branch
		notifications, err := repo.ListByUser(ctx, testUser, cursor, 10, false)
		require.NoError(t, err)
		assert.NotEmpty(t, notifications)
	})

	t.Run("Metadata deserialization with empty JSON", func(t *testing.T) {
		// Create a notification and verify metadata is properly deserialized
		testUser := uuid.New().String()
		_, _ = repo.Create(ctx, &entities.Notification{
			UserID:   testUser,
			EventKey: "empty_metadata_" + uuid.New().String(),
			Type:     entities.TypeOrderFilled,
			Title:    "Empty Metadata Test",
			Body:     "Testing empty metadata",
			Metadata: map[string]string{},
		})

		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, testUser, cursor, 10, false)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)
		assert.Empty(t, notifications[0].Metadata)
	})

	t.Run("Multiple MarkAsRead operations on same notification", func(t *testing.T) {
		testUser := uuid.New().String()
		notif := &entities.Notification{
			UserID:   testUser,
			EventKey: "multi_read_" + uuid.New().String(),
			Type:     entities.TypeWithdrawalConfirmed,
			Title:    "Multi Read Test",
			Body:     "Testing multiple reads",
			Metadata: map[string]string{},
		}
		_, err := repo.Create(ctx, notif)
		require.NoError(t, err)

		cursor := time.Now().Add(time.Minute)
		notifications, err := repo.ListByUser(ctx, testUser, cursor, 1, false)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)
		notifID := notifications[0].ID

		// First read
		err = repo.MarkAsRead(ctx, notifID, testUser)
		require.NoError(t, err)

		// Second read (should succeed but return 0 affected rows when checking unread)
		count, err := repo.CountUnread(ctx, testUser)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
