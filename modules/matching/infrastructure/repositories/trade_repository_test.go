package repositories

import (
	"context"
	"testing"

	"booker/modules/matching/domain"
	"booker/modules/matching/domain/entities"
	tc "booker/test/testcontainers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestUser inserts a user directly for FK constraint, returns UUID.
func createTestUser(t *testing.T, ctx context.Context, containers *tc.TestContainers) string {
	t.Helper()
	id := uuid.New().String()
	_, err := containers.Database.Exec(ctx,
		"INSERT INTO users (id, email, password, role) VALUES ($1, $2, 'hash', 'user')",
		id, id+"@test.com",
	)
	require.NoError(t, err)
	return id
}

func seedAssetsAndPairs(t *testing.T, ctx context.Context, containers *tc.TestContainers) {
	t.Helper()
	_, err := containers.Database.Exec(
		ctx,
		"INSERT INTO assets (id, name, decimals) VALUES ('BTC', 'Bitcoin', 8), ('USDT', 'Tether', 6) ON CONFLICT DO NOTHING",
	)
	require.NoError(t, err)
	_, err = containers.Database.Exec(
		ctx,
		"INSERT INTO trading_pairs (id, base_asset, quote_asset, status, min_qty, tick_size) VALUES ('BTC_USDT', 'BTC', 'USDT', 'active', 0.00001, 0.01) ON CONFLICT DO NOTHING",
	)
	require.NoError(t, err)
}

func createTestOrder(t *testing.T, ctx context.Context, containers *tc.TestContainers, userID, pairID, side string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := containers.Database.Exec(ctx,
		"INSERT INTO orders (id, user_id, pair_id, side, type, price, quantity, filled_qty, status) VALUES ($1, $2, $3, $4, 'limit', 50000, 1, 0, 'new')",
		id, userID, pairID, side,
	)
	require.NoError(t, err)
	return id
}

func TestTradeRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	repo := NewTradeRepository(containers.Database)
	ctx := context.Background()

	seedAssetsAndPairs(t, ctx, containers)
	buyerID := createTestUser(t, ctx, containers)
	sellerID := createTestUser(t, ctx, containers)
	buyOrderID := createTestOrder(t, ctx, containers, buyerID, "BTC_USDT", "buy")
	sellOrderID := createTestOrder(t, ctx, containers, sellerID, "BTC_USDT", "sell")

	var createdTradeID string

	t.Run("Create and GetByID", func(t *testing.T) {
		trade := &entities.Trade{
			PairID:      "BTC_USDT",
			BuyOrderID:  buyOrderID,
			SellOrderID: sellOrderID,
			Price:       decimal.NewFromFloat(50000),
			Quantity:    decimal.NewFromFloat(0.5),
			BuyerID:     buyerID,
			SellerID:    sellerID,
		}
		created, err := repo.Create(ctx, trade)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "BTC_USDT", created.PairID)
		assert.Equal(t, buyOrderID, created.BuyOrderID)
		assert.Equal(t, sellOrderID, created.SellOrderID)
		assert.True(t, created.Price.Equal(decimal.NewFromFloat(50000)))
		assert.True(t, created.Quantity.Equal(decimal.NewFromFloat(0.5)))
		assert.Equal(t, buyerID, created.BuyerID)
		assert.Equal(t, sellerID, created.SellerID)
		assert.False(t, created.ExecutedAt.IsZero())

		createdTradeID = created.ID

		// GetByID
		found, err := repo.GetByID(ctx, createdTradeID)
		require.NoError(t, err)
		assert.Equal(t, createdTradeID, found.ID)
		assert.Equal(t, "BTC_USDT", found.PairID)
		assert.True(t, found.Price.Equal(decimal.NewFromFloat(50000)))
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New().String())
		assert.Equal(t, domain.ErrTradeNotFound, err)
	})

	t.Run("ListByPair", func(t *testing.T) {
		// Create a second trade with new orders
		buyOrder2 := createTestOrder(t, ctx, containers, buyerID, "BTC_USDT", "buy")
		sellOrder2 := createTestOrder(t, ctx, containers, sellerID, "BTC_USDT", "sell")
		_, err := repo.Create(ctx, &entities.Trade{
			PairID:      "BTC_USDT",
			BuyOrderID:  buyOrder2,
			SellOrderID: sellOrder2,
			Price:       decimal.NewFromFloat(51000),
			Quantity:    decimal.NewFromFloat(0.3),
			BuyerID:     buyerID,
			SellerID:    sellerID,
		})
		require.NoError(t, err)

		trades, err := repo.ListByPair(ctx, "BTC_USDT", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(trades), 2)
		for _, tr := range trades {
			assert.Equal(t, "BTC_USDT", tr.PairID)
		}
	})

	t.Run("ListByPair with pagination", func(t *testing.T) {
		trades1, err := repo.ListByPair(ctx, "BTC_USDT", 1, 0)
		require.NoError(t, err)
		assert.Len(t, trades1, 1)

		trades2, err := repo.ListByPair(ctx, "BTC_USDT", 1, 1)
		require.NoError(t, err)
		assert.Len(t, trades2, 1)
		assert.NotEqual(t, trades1[0].ID, trades2[0].ID)
	})

	t.Run("ListByPair empty result", func(t *testing.T) {
		trades, err := repo.ListByPair(ctx, "ETH_USDT", 10, 0)
		require.NoError(t, err)
		assert.Empty(t, trades)
	})

	t.Run("ListByPair with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.ListByPair(cancelCtx, "BTC_USDT", 10, 0)
		assert.Error(t, err)
	})

	t.Run("GetByID with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.GetByID(cancelCtx, createdTradeID)
		assert.Error(t, err)
	})
}
