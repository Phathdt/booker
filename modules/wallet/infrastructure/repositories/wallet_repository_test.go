package repositories

import (
	"context"
	"testing"

	"booker/modules/wallet/domain"
	tc "booker/test/testcontainers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestUser inserts a user directly for FK constraint, returns UUID.
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

func TestWalletRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	repo := NewWalletRepository(containers.Database)
	ctx := context.Background()

	// Seed test assets
	_, _ = containers.Database.Exec(ctx, "INSERT INTO assets (id, name, decimals) VALUES ('USDT', 'Tether', 6), ('BTC', 'Bitcoin', 8) ON CONFLICT DO NOTHING")

	userID := createTestUser(t, containers)

	t.Run("GetOrCreate creates wallet", func(t *testing.T) {
		w, err := repo.GetOrCreate(ctx, userID, "USDT")
		require.NoError(t, err)
		assert.Equal(t, userID, w.UserID)
		assert.Equal(t, "USDT", w.AssetID)
		assert.True(t, w.Available.Equal(decimal.Zero))
		assert.True(t, w.Locked.Equal(decimal.Zero))
	})

	t.Run("GetOrCreate returns existing", func(t *testing.T) {
		w1, _ := repo.GetOrCreate(ctx, userID, "USDT")
		w2, _ := repo.GetOrCreate(ctx, userID, "USDT")
		assert.Equal(t, w1.ID, w2.ID)
	})

	t.Run("Deposit", func(t *testing.T) {
		w, err := repo.Deposit(ctx, userID, "USDT", decimal.NewFromFloat(1000))
		require.NoError(t, err)
		assert.True(t, w.Available.Equal(decimal.NewFromFloat(1000)))
	})

	t.Run("Hold", func(t *testing.T) {
		w, err := repo.Hold(ctx, userID, "USDT", decimal.NewFromFloat(300))
		require.NoError(t, err)
		assert.True(t, w.Available.Equal(decimal.NewFromFloat(700)))
		assert.True(t, w.Locked.Equal(decimal.NewFromFloat(300)))
	})

	t.Run("Hold insufficient balance", func(t *testing.T) {
		_, err := repo.Hold(ctx, userID, "USDT", decimal.NewFromFloat(9999))
		assert.Equal(t, domain.ErrInsufficientBalance, err)
	})

	t.Run("Release", func(t *testing.T) {
		w, err := repo.Release(ctx, userID, "USDT", decimal.NewFromFloat(100))
		require.NoError(t, err)
		assert.True(t, w.Available.Equal(decimal.NewFromFloat(800)))
		assert.True(t, w.Locked.Equal(decimal.NewFromFloat(200)))
	})

	t.Run("Release insufficient locked", func(t *testing.T) {
		_, err := repo.Release(ctx, userID, "USDT", decimal.NewFromFloat(9999))
		assert.Equal(t, domain.ErrInsufficientLocked, err)
	})

	t.Run("Settle", func(t *testing.T) {
		w, err := repo.Settle(ctx, userID, "USDT", decimal.NewFromFloat(200))
		require.NoError(t, err)
		assert.True(t, w.Available.Equal(decimal.NewFromFloat(800)))
		assert.True(t, w.Locked.Equal(decimal.Zero))
	})

	t.Run("Withdraw", func(t *testing.T) {
		w, err := repo.Withdraw(ctx, userID, "USDT", decimal.NewFromFloat(300))
		require.NoError(t, err)
		assert.True(t, w.Available.Equal(decimal.NewFromFloat(500)))
	})

	t.Run("Withdraw insufficient", func(t *testing.T) {
		_, err := repo.Withdraw(ctx, userID, "USDT", decimal.NewFromFloat(9999))
		assert.Equal(t, domain.ErrInsufficientBalance, err)
	})

	t.Run("GetByUserID", func(t *testing.T) {
		// Create a second wallet
		_, _ = repo.GetOrCreate(ctx, userID, "BTC")
		wallets, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(wallets), 2)
	})
}
