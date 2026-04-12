package repositories

import (
	"context"
	"testing"

	"booker/modules/order/domain"
	"booker/modules/order/domain/entities"
	tc "booker/test/testcontainers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, containers *tc.TestContainers) string {
	t.Helper()
	id := uuid.New().String()
	_, err := containers.Database.Exec(context.Background(),
		"INSERT INTO users (id, email, password, role) VALUES ($1, $2, 'hash', 'user')",
		id, id+"@test.com",
	)
	require.NoError(t, err)
	return id
}

func seedAssetsAndPairs(t *testing.T, containers *tc.TestContainers) {
	t.Helper()
	ctx := context.Background()
	_, _ = containers.Database.Exec(
		ctx,
		"INSERT INTO assets (id, name, decimals) VALUES ('BTC', 'Bitcoin', 8), ('USDT', 'Tether', 6) ON CONFLICT DO NOTHING",
	)
	_, _ = containers.Database.Exec(
		ctx,
		"INSERT INTO trading_pairs (id, base_asset, quote_asset, status, min_qty, tick_size) VALUES ('BTC_USDT', 'BTC', 'USDT', 'active', 0.00001, 0.01) ON CONFLICT DO NOTHING",
	)
}

func TestOrderRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	repo := NewOrderRepository(containers.Database)
	ctx := context.Background()

	seedAssetsAndPairs(t, containers)
	userID := createTestUser(t, containers)

	var createdOrderID string

	t.Run("GetTradingPair", func(t *testing.T) {
		pair, err := repo.GetTradingPair(ctx, "BTC_USDT")
		require.NoError(t, err)
		assert.Equal(t, "BTC", pair.BaseAsset)
		assert.Equal(t, "USDT", pair.QuoteAsset)
		assert.Equal(t, "active", pair.Status)
	})

	t.Run("GetTradingPair not found", func(t *testing.T) {
		_, err := repo.GetTradingPair(ctx, "INVALID_PAIR")
		assert.Equal(t, domain.ErrPairNotFound, err)
	})

	t.Run("Create + GetByID", func(t *testing.T) {
		order := &entities.Order{
			UserID:   userID,
			PairID:   "BTC_USDT",
			Side:     "buy",
			Type:     "limit",
			Price:    decimal.NewFromFloat(50000),
			Quantity: decimal.NewFromFloat(0.5),
		}
		created, err := repo.Create(ctx, order)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "new", created.Status)
		assert.True(t, created.FilledQty.Equal(decimal.Zero))

		createdOrderID = created.ID

		fetched, err := repo.GetByID(ctx, createdOrderID)
		require.NoError(t, err)
		assert.Equal(t, createdOrderID, fetched.ID)
		assert.Equal(t, userID, fetched.UserID)
	})

	t.Run("GetByIDAndUser", func(t *testing.T) {
		order, err := repo.GetByIDAndUser(ctx, createdOrderID, userID)
		require.NoError(t, err)
		assert.Equal(t, createdOrderID, order.ID)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New().String())
		assert.Equal(t, domain.ErrOrderNotFound, err)
	})

	t.Run("GetByIDAndUser wrong user", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := repo.GetByIDAndUser(ctx, createdOrderID, otherUserID)
		assert.Equal(t, domain.ErrOrderNotFound, err)
	})

	t.Run("List by user", func(t *testing.T) {
		// Create a second order
		_, _ = repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "sell", Type: "limit",
			Price: decimal.NewFromFloat(55000), Quantity: decimal.NewFromFloat(1),
		})

		orders, err := repo.List(ctx, userID, nil, nil, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(orders), 2)
	})

	t.Run("List by pair filter", func(t *testing.T) {
		pairID := "BTC_USDT"
		orders, err := repo.List(ctx, userID, &pairID, nil, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(orders), 2)
		for _, o := range orders {
			assert.Equal(t, "BTC_USDT", o.PairID)
		}
	})

	t.Run("List by status filter", func(t *testing.T) {
		status := "new"
		orders, err := repo.List(ctx, userID, nil, &status, 10, 0)
		require.NoError(t, err)
		for _, o := range orders {
			assert.Equal(t, "new", o.Status)
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		orders, err := repo.List(ctx, userID, nil, nil, 1, 0)
		require.NoError(t, err)
		assert.Len(t, orders, 1)

		orders2, err := repo.List(ctx, userID, nil, nil, 1, 1)
		require.NoError(t, err)
		assert.Len(t, orders2, 1)
		assert.NotEqual(t, orders[0].ID, orders2[0].ID)
	})

	t.Run("List empty result", func(t *testing.T) {
		otherUser := createTestUser(t, containers)
		orders, err := repo.List(ctx, otherUser, nil, nil, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, orders)
	})

	t.Run("Cancel order", func(t *testing.T) {
		cancelled, err := repo.Cancel(ctx, createdOrderID, userID)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", cancelled.Status)
	})

	t.Run("Cancel already cancelled", func(t *testing.T) {
		_, err := repo.Cancel(ctx, createdOrderID, userID)
		assert.Equal(t, domain.ErrOrderNotCancellable, err)
	})

	t.Run("UpdateFilledQty", func(t *testing.T) {
		// Create a fresh order for fill test
		fresh, err := repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		})
		require.NoError(t, err)

		updated, err := repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(0.5), "partial")
		require.NoError(t, err)
		assert.Equal(t, "partial", updated.Status)
		assert.True(t, updated.FilledQty.Equal(decimal.NewFromFloat(0.5)))

		// Fill completely
		filled, err := repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(1), "filled")
		require.NoError(t, err)
		assert.Equal(t, "filled", filled.Status)
	})

	t.Run("UpdateFilledQty on cancelled order", func(t *testing.T) {
		// createdOrderID was cancelled earlier
		_, err := repo.UpdateFilledQty(ctx, createdOrderID, decimal.NewFromFloat(0.1), "partial")
		assert.Equal(t, domain.ErrOrderNotFillable, err)
	})

	t.Run("UpdateFilledQty backward fill rejected", func(t *testing.T) {
		fresh, err := repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		})
		require.NoError(t, err)

		_, err = repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(0.5), "partial")
		require.NoError(t, err)

		// Try backward fill (0.3 < current 0.5) — should fail
		_, err = repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(0.3), "partial")
		assert.Equal(t, domain.ErrOrderNotFillable, err)
	})

	t.Run("UpdateFilledQty exceeds quantity", func(t *testing.T) {
		fresh, err := repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		})
		require.NoError(t, err)

		_, err = repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(2), "filled")
		assert.Equal(t, domain.ErrOrderNotFillable, err)
	})

	t.Run("GetTradingPair with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.GetTradingPair(cancelCtx, "BTC_USDT")
		assert.Error(t, err)
	})

	t.Run("Create with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Create(cancelCtx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		})
		assert.Error(t, err)
	})

	t.Run("GetByID with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.GetByID(cancelCtx, createdOrderID)
		assert.Error(t, err)
	})

	t.Run("GetByIDAndUser with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.GetByIDAndUser(cancelCtx, createdOrderID, userID)
		assert.Error(t, err)
	})

	t.Run("List with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.List(cancelCtx, userID, nil, nil, 10, 0)
		assert.Error(t, err)
	})

	t.Run("Cancel with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Cancel(cancelCtx, createdOrderID, userID)
		assert.Error(t, err)
	})

	t.Run("UpdateFilledQty with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.UpdateFilledQty(cancelCtx, createdOrderID, decimal.NewFromFloat(0.5), "partial")
		assert.Error(t, err)
	})

	t.Run("Cancel order not found", func(t *testing.T) {
		_, err := repo.Cancel(ctx, uuid.New().String(), userID)
		assert.Equal(t, domain.ErrOrderNotCancellable, err)
	})

	t.Run("Cancel by wrong user", func(t *testing.T) {
		wrongUser := createTestUser(t, containers)
		fresh, err := repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(1),
		})
		require.NoError(t, err)

		_, err = repo.Cancel(ctx, fresh.ID, wrongUser)
		assert.Equal(t, domain.ErrOrderNotCancellable, err)
	})

	t.Run("UpdateFilledQty not found", func(t *testing.T) {
		_, err := repo.UpdateFilledQty(ctx, uuid.New().String(), decimal.NewFromFloat(0.5), "partial")
		assert.Equal(t, domain.ErrOrderNotFillable, err)
	})

	t.Run("List multiple filters combined", func(t *testing.T) {
		// Create another pair and orders with different statuses
		// First ensure ETH asset exists
		_, _ = containers.Database.Exec(ctx,
			"INSERT INTO assets (id, name, decimals) VALUES ('ETH', 'Ethereum', 18) ON CONFLICT DO NOTHING",
		)
		_, err := containers.Database.Exec(ctx,
			"INSERT INTO trading_pairs (id, base_asset, quote_asset, status, min_qty, tick_size) VALUES ('ETH_USDT', 'ETH', 'USDT', 'active', 0.0001, 0.01) ON CONFLICT DO NOTHING",
		)
		require.NoError(t, err)

		ethOrder, err := repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "ETH_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(3000), Quantity: decimal.NewFromFloat(1),
		})
		require.NoError(t, err)

		pairID := "ETH_USDT"
		status := "new"
		orders, err := repo.List(ctx, userID, &pairID, &status, 10, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, orders)
		for _, o := range orders {
			assert.Equal(t, "ETH_USDT", o.PairID)
			assert.Equal(t, "new", o.Status)
		}

		// Cancel the ETH order
		_, _ = repo.Cancel(ctx, ethOrder.ID, userID)

		// Filter for cancelled should only show cancelled
		cancelledStatus := "cancelled"
		orders, err = repo.List(ctx, userID, &pairID, &cancelledStatus, 10, 0)
		require.NoError(t, err)
		for _, o := range orders {
			assert.Equal(t, "cancelled", o.Status)
		}
	})

	t.Run("Create multiple orders and list with limit", func(t *testing.T) {
		newUser := createTestUser(t, containers)
		// Create 5 orders
		for i := 0; i < 5; i++ {
			_, err := repo.Create(ctx, &entities.Order{
				UserID: newUser, PairID: "BTC_USDT", Side: "buy", Type: "limit",
				Price:    decimal.NewFromFloat(50000 + float64(i*100)),
				Quantity: decimal.NewFromFloat(0.1),
			})
			require.NoError(t, err)
		}

		// Test pagination with limit 2
		orders1, err := repo.List(ctx, newUser, nil, nil, 2, 0)
		require.NoError(t, err)
		assert.Len(t, orders1, 2)

		orders2, err := repo.List(ctx, newUser, nil, nil, 2, 2)
		require.NoError(t, err)
		assert.Len(t, orders2, 2)

		// Verify they are different orders
		assert.NotEqual(t, orders1[0].ID, orders2[0].ID)
	})

	t.Run("UpdateFilledQty partial fills", func(t *testing.T) {
		fresh, err := repo.Create(ctx, &entities.Order{
			UserID: userID, PairID: "BTC_USDT", Side: "buy", Type: "limit",
			Price: decimal.NewFromFloat(50000), Quantity: decimal.NewFromFloat(10),
		})
		require.NoError(t, err)

		// First partial fill
		filled1, err := repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(3), "partial")
		require.NoError(t, err)
		assert.Equal(t, "partial", filled1.Status)
		assert.True(t, filled1.FilledQty.Equal(decimal.NewFromFloat(3)))

		// Second partial fill
		filled2, err := repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(7), "partial")
		require.NoError(t, err)
		assert.Equal(t, "partial", filled2.Status)
		assert.True(t, filled2.FilledQty.Equal(decimal.NewFromFloat(7)))

		// Complete fill
		filled3, err := repo.UpdateFilledQty(ctx, fresh.ID, decimal.NewFromFloat(10), "filled")
		require.NoError(t, err)
		assert.Equal(t, "filled", filled3.Status)
		assert.True(t, filled3.FilledQty.Equal(decimal.NewFromFloat(10)))
	})
}
