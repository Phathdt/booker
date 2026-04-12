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
	_, _ = containers.Database.Exec(
		ctx,
		"INSERT INTO assets (id, name, decimals) VALUES ('USDT', 'Tether', 6), ('BTC', 'Bitcoin', 8) ON CONFLICT DO NOTHING",
	)

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

	t.Run("GetByUserID empty for unknown user", func(t *testing.T) {
		unknownUserID := uuid.New().String()
		wallets, err := repo.GetByUserID(ctx, unknownUserID)
		require.NoError(t, err)
		assert.Empty(t, wallets)
	})

	t.Run("Deposit with create-then-retry path", func(t *testing.T) {
		// Test the fast path (wallet exists)
		newUserID := createTestUser(t, containers)
		// First deposit will create wallet (through GetOrCreate) then deposit
		w1, err := repo.Deposit(ctx, newUserID, "USDT", decimal.NewFromFloat(500))
		require.NoError(t, err)
		assert.True(t, w1.Available.Equal(decimal.NewFromFloat(500)))

		// Second deposit should use fast path
		w2, err := repo.Deposit(ctx, newUserID, "USDT", decimal.NewFromFloat(300))
		require.NoError(t, err)
		assert.True(t, w2.Available.Equal(decimal.NewFromFloat(800)))
	})

	t.Run("GetOrCreate with nonexistent wallet for user without profile", func(t *testing.T) {
		// Create a fresh user
		freshUserID := createTestUser(t, containers)
		w, err := repo.GetOrCreate(ctx, freshUserID, "USDT")
		require.NoError(t, err)
		assert.Equal(t, freshUserID, w.UserID)
		assert.Equal(t, "USDT", w.AssetID)
		assert.True(t, w.Available.Equal(decimal.Zero))
		assert.True(t, w.Locked.Equal(decimal.Zero))
	})

	t.Run("Settle with error path", func(t *testing.T) {
		settleUserID := createTestUser(t, containers)
		// Try to settle on non-existent wallet
		_, err := repo.Settle(ctx, settleUserID, "USDT", decimal.NewFromFloat(100))
		assert.Equal(t, domain.ErrInsufficientLocked, err)
	})

	t.Run("Hold then Release complete cycle", func(t *testing.T) {
		cycleUserID := createTestUser(t, containers)
		// Deposit
		w1, _ := repo.Deposit(ctx, cycleUserID, "USDT", decimal.NewFromFloat(1000))
		assert.True(t, w1.Available.Equal(decimal.NewFromFloat(1000)))

		// Hold
		w2, _ := repo.Hold(ctx, cycleUserID, "USDT", decimal.NewFromFloat(250))
		assert.True(t, w2.Available.Equal(decimal.NewFromFloat(750)))
		assert.True(t, w2.Locked.Equal(decimal.NewFromFloat(250)))

		// Release
		w3, _ := repo.Release(ctx, cycleUserID, "USDT", decimal.NewFromFloat(250))
		assert.True(t, w3.Available.Equal(decimal.NewFromFloat(1000)))
		assert.True(t, w3.Locked.Equal(decimal.Zero))
	})

	t.Run("Release insufficient locked error", func(t *testing.T) {
		releaseUserID := createTestUser(t, containers)
		_, _ = repo.Deposit(ctx, releaseUserID, "USDT", decimal.NewFromFloat(1000))
		// Try to release without any hold
		_, err := repo.Release(ctx, releaseUserID, "USDT", decimal.NewFromFloat(1))
		assert.Equal(t, domain.ErrInsufficientLocked, err)
	})

	t.Run("Withdraw insufficient balance error", func(t *testing.T) {
		withdrawUserID := createTestUser(t, containers)
		// Try withdraw without deposit
		_, err := repo.Withdraw(ctx, withdrawUserID, "USDT", decimal.NewFromFloat(1))
		assert.Equal(t, domain.ErrInsufficientBalance, err)
	})

	t.Run("GetOrCreate with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.GetOrCreate(cancelCtx, userID, "USDT")
		assert.Error(t, err)
	})

	t.Run("GetByUserID with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.GetByUserID(cancelCtx, userID)
		assert.Error(t, err)
	})

	t.Run("Deposit with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Deposit(cancelCtx, userID, "USDT", decimal.NewFromFloat(100))
		assert.Error(t, err)
	})

	t.Run("Withdraw with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Withdraw(cancelCtx, userID, "USDT", decimal.NewFromFloat(100))
		assert.Error(t, err)
	})

	t.Run("Hold with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Hold(cancelCtx, userID, "USDT", decimal.NewFromFloat(100))
		assert.Error(t, err)
	})

	t.Run("Release with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Release(cancelCtx, userID, "USDT", decimal.NewFromFloat(100))
		assert.Error(t, err)
	})

	t.Run("Settle with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := repo.Settle(cancelCtx, userID, "USDT", decimal.NewFromFloat(100))
		assert.Error(t, err)
	})
}
