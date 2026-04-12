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

func TestQueries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	queries := NewQueries(containers.Database)
	ctx := context.Background()

	seedAssetsAndPairs(t, ctx, containers)

	t.Run("ListActiveTradingPairs returns active pairs", func(t *testing.T) {
		pairs, err := queries.ListActiveTradingPairs(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, pairs)

		for _, pair := range pairs {
			assert.Equal(t, "active", pair.Status)
		}
	})

	t.Run("ListActiveTradingPairs with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := queries.ListActiveTradingPairs(cancelCtx)
		assert.Error(t, err)
	})

	t.Run("ListOpenOrdersByPair returns open orders", func(t *testing.T) {
		userID := createTestUser(t, ctx, containers)
		// Create some open orders
		for i := 0; i < 3; i++ {
			_, err := containers.Database.Exec(ctx,
				"INSERT INTO orders (id, user_id, pair_id, side, type, price, quantity, filled_qty, status) VALUES ($1, $2, 'BTC_USDT', 'buy', 'limit', 50000, 1, 0, 'new')",
				uuid.New().String(), userID,
			)
			require.NoError(t, err)
		}

		orders, err := queries.ListOpenOrdersByPair(ctx, "BTC_USDT")
		require.NoError(t, err)
		assert.NotEmpty(t, orders)

		for _, order := range orders {
			assert.Equal(t, "BTC_USDT", order.PairID)
			// Open orders should be in 'new' or 'partial' status
			assert.True(t, order.Status == "new" || order.Status == "partial")
		}
	})

	t.Run("ListOpenOrdersByPair with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := queries.ListOpenOrdersByPair(cancelCtx, "BTC_USDT")
		assert.Error(t, err)
	})

	t.Run("ListOpenOrdersByPair for non-existent pair", func(t *testing.T) {
		orders, err := queries.ListOpenOrdersByPair(ctx, "NONEXISTENT_PAIR")
		require.NoError(t, err)
		assert.Empty(t, orders) // No orders for this pair, should be empty
	})

	t.Run("ListOpenOrdersByPair with multiple pairs", func(t *testing.T) {
		// Insert another pair — first ensure ETH asset exists
		_, _ = containers.Database.Exec(ctx,
			"INSERT INTO assets (id, name, decimals) VALUES ('ETH', 'Ethereum', 18) ON CONFLICT DO NOTHING",
		)
		_, err := containers.Database.Exec(ctx,
			"INSERT INTO trading_pairs (id, base_asset, quote_asset, status, min_qty, tick_size) VALUES ('ETH_USDT', 'ETH', 'USDT', 'active', 0.0001, 0.01) ON CONFLICT DO NOTHING",
		)
		require.NoError(t, err)

		userID := createTestUser(t, ctx, containers)

		// Create orders for BTC_USDT
		_, _ = containers.Database.Exec(ctx,
			"INSERT INTO orders (id, user_id, pair_id, side, type, price, quantity, filled_qty, status) VALUES ($1, $2, 'BTC_USDT', 'buy', 'limit', 50000, 1, 0, 'new')",
			uuid.New().String(), userID,
		)

		// Create orders for ETH_USDT
		_, _ = containers.Database.Exec(ctx,
			"INSERT INTO orders (id, user_id, pair_id, side, type, price, quantity, filled_qty, status) VALUES ($1, $2, 'ETH_USDT', 'sell', 'limit', 3000, 1, 0, 'new')",
			uuid.New().String(), userID,
		)

		// List should only return the requested pair
		btcOrders, err := queries.ListOpenOrdersByPair(ctx, "BTC_USDT")
		require.NoError(t, err)
		for _, o := range btcOrders {
			assert.Equal(t, "BTC_USDT", o.PairID)
		}

		ethOrders, err := queries.ListOpenOrdersByPair(ctx, "ETH_USDT")
		require.NoError(t, err)
		for _, o := range ethOrders {
			assert.Equal(t, "ETH_USDT", o.PairID)
		}
	})
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

	t.Run("Create with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Create(cancelCtx, &entities.Trade{
			PairID:      "BTC_USDT",
			BuyOrderID:  buyOrderID,
			SellOrderID: sellOrderID,
			Price:       decimal.NewFromFloat(50000),
			Quantity:    decimal.NewFromFloat(0.5),
			BuyerID:     buyerID,
			SellerID:    sellerID,
		})
		assert.Error(t, err)
	})

	t.Run("Create with database error", func(t *testing.T) {
		// Try to create a trade with non-existent orders
		_, err := repo.Create(ctx, &entities.Trade{
			PairID:      "BTC_USDT",
			BuyOrderID:  "invalid-order-id",
			SellOrderID: "invalid-order-id",
			Price:       decimal.NewFromFloat(50000),
			Quantity:    decimal.NewFromFloat(0.5),
			BuyerID:     buyerID,
			SellerID:    sellerID,
		})
		assert.Error(t, err) // Should fail due to foreign key constraint
	})

	t.Run("Create multiple trades and verify all returned", func(t *testing.T) {
		// Create a new user pair for isolation
		trader1 := createTestUser(t, ctx, containers)
		trader2 := createTestUser(t, ctx, containers)

		// Create multiple orders for trading
		orders := make([]string, 4) // 2 buy, 2 sell
		for i := 0; i < 4; i++ {
			if i%2 == 0 {
				orders[i] = createTestOrder(t, ctx, containers, trader1, "BTC_USDT", "buy")
			} else {
				orders[i] = createTestOrder(t, ctx, containers, trader2, "BTC_USDT", "sell")
			}
		}

		// Create multiple trades
		for i := 0; i < 2; i++ {
			buyOrderID := orders[i*2]
			sellOrderID := orders[i*2+1]
			_, err := repo.Create(ctx, &entities.Trade{
				PairID:      "BTC_USDT",
				BuyOrderID:  buyOrderID,
				SellOrderID: sellOrderID,
				Price:       decimal.NewFromFloat(50000 + float64(i*100)),
				Quantity:    decimal.NewFromFloat(0.25),
				BuyerID:     trader1,
				SellerID:    trader2,
			})
			require.NoError(t, err)
		}

		// Verify all trades are returned
		trades, err := repo.ListByPair(ctx, "BTC_USDT", 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(trades), 4) // At least our 2 + original 2
	})

	t.Run("GetByID not found with different ID format", func(t *testing.T) {
		// Test with invalid UUID format
		_, err := repo.GetByID(ctx, "not-a-uuid")
		assert.Error(t, err)
	})

	t.Run("ListByPair pagination with exact boundary", func(t *testing.T) {
		// Create exactly 3 trades for this pair
		pair := "BTC_USDT"
		buyer := createTestUser(t, ctx, containers)
		seller := createTestUser(t, ctx, containers)

		for i := 0; i < 3; i++ {
			buyOrder := createTestOrder(t, ctx, containers, buyer, pair, "buy")
			sellOrder := createTestOrder(t, ctx, containers, seller, pair, "sell")
			_, err := repo.Create(ctx, &entities.Trade{
				PairID:      pair,
				BuyOrderID:  buyOrder,
				SellOrderID: sellOrder,
				Price:       decimal.NewFromFloat(50000),
				Quantity:    decimal.NewFromFloat(0.1),
				BuyerID:     buyer,
				SellerID:    seller,
			})
			require.NoError(t, err)
		}

		// Test limit = total count
		trades1, err := repo.ListByPair(ctx, pair, 3, 0)
		require.NoError(t, err)
		assert.Equal(t, 3, len(trades1))

		// Test offset beyond results
		trades2, err := repo.ListByPair(ctx, pair, 10, 10)
		require.NoError(t, err)
		assert.Empty(t, trades2)
	})

	t.Run("Trade data integrity verification", func(t *testing.T) {
		// Create a specific trade and verify all fields are preserved
		trade := &entities.Trade{
			PairID:      "BTC_USDT",
			BuyOrderID:  buyOrderID,
			SellOrderID: sellOrderID,
			Price:       decimal.NewFromFloat(52500),
			Quantity:    decimal.NewFromFloat(0.75),
			BuyerID:     buyerID,
			SellerID:    sellerID,
		}
		created, err := repo.Create(ctx, trade)
		require.NoError(t, err)

		// Retrieve and verify exact field matching
		retrieved, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, "BTC_USDT", retrieved.PairID)
		assert.Equal(t, buyOrderID, retrieved.BuyOrderID)
		assert.Equal(t, sellOrderID, retrieved.SellOrderID)
		assert.True(t, retrieved.Price.Equal(decimal.NewFromFloat(52500)))
		assert.True(t, retrieved.Quantity.Equal(decimal.NewFromFloat(0.75)))
		assert.Equal(t, buyerID, retrieved.BuyerID)
		assert.Equal(t, sellerID, retrieved.SellerID)
		assert.NotZero(t, retrieved.ExecutedAt)
	})
}
