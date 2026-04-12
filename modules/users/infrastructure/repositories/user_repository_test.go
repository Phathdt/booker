package repositories

import (
	"context"
	"fmt"
	"testing"

	tc "booker/test/testcontainers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	repo := NewUserRepository(containers.Database)
	ctx := context.Background()

	t.Run("Create and GetByID", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
		user, err := repo.Create(ctx, "create-test@example.com", string(hash), "user")

		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "create-test@example.com", user.Email)
		assert.Equal(t, "user", user.Role)
		assert.Equal(t, "active", user.Status)

		// GetByID
		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
		created, err := repo.Create(ctx, "email-test@example.com", string(hash), "user")
		require.NoError(t, err)

		found, err := repo.GetByEmail(ctx, "email-test@example.com")
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
	})

	t.Run("GetByEmail not found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		users, total, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(users), 2)
	})

	t.Run("Create duplicate email", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
		_, err := repo.Create(ctx, "create-test@example.com", string(hash), "user")
		assert.Error(t, err) // unique constraint violation
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
		assert.Error(t, err)
	})

	t.Run("Update nonexistent", func(t *testing.T) {
		_, err := repo.Update(ctx, "00000000-0000-0000-0000-000000000000", "x@x.com", "user", "active")
		assert.Error(t, err)
	})

	t.Run("List with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, _, err := repo.List(cancelCtx, 10, 0)
		assert.Error(t, err)
	})

	t.Run("List returns total count", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 5; i++ {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
			_, _ = repo.Create(ctx, fmt.Sprintf("list-test-%d@example.com", i), string(hash), "user")
		}

		users, total, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(5))
		assert.Len(t, users, 2) // limit is 2
	})

	t.Run("List with offset", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
		_, _ = repo.Create(ctx, "offset-test-1@example.com", string(hash), "user")
		_, _ = repo.Create(ctx, "offset-test-2@example.com", string(hash), "user")

		// First page
		page1, total, err := repo.List(ctx, 1, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.Len(t, page1, 1)

		// Second page
		page2, _, err := repo.List(ctx, 1, 1)
		require.NoError(t, err)
		assert.Len(t, page2, 1)

		// Verify different users
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("Update", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
		user, err := repo.Create(ctx, "update-test@example.com", string(hash), "user")
		require.NoError(t, err)

		updated, err := repo.Update(ctx, user.ID, "updated@example.com", "admin", "inactive")
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", updated.Email)
		assert.Equal(t, "admin", updated.Role)
		assert.Equal(t, "inactive", updated.Status)
	})
}
