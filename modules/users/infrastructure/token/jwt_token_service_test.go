package token

import (
	"context"
	"testing"
	"time"

	"booker/config"
	tc "booker/test/testcontainers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTokenService(t *testing.T) *jwtTokenService {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	cfg := config.JWTConfig{
		Secret:     "test-secret-key-for-jwt",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := NewJWTTokenService(containers.RedisClient, cfg)
	return svc.(*jwtTokenService)
}

func TestJWTTokenService_GenerateAndValidateAccess(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	access, refresh, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	// Validate access token
	claims, err := svc.ValidateAccessToken(ctx, access)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "user", claims.Role)
	assert.NotEmpty(t, claims.JTI)
}

func TestJWTTokenService_ValidateRefreshToken_SingleUse(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	_, refresh, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	// First validation succeeds
	claims, err := svc.ValidateRefreshToken(ctx, refresh)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)

	// Second validation fails (single-use, JTI deleted)
	_, err = svc.ValidateRefreshToken(ctx, refresh)
	assert.Error(t, err)
}

func TestJWTTokenService_RevokeAllUserTokens(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	access, _, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	// Access token works before revoke
	_, err = svc.ValidateAccessToken(ctx, access)
	require.NoError(t, err)

	// Revoke all tokens
	err = svc.RevokeAllUserTokens(ctx, "user-1")
	require.NoError(t, err)

	// Access token fails after revoke
	_, err = svc.ValidateAccessToken(ctx, access)
	assert.Error(t, err)
}

func TestJWTTokenService_InvalidToken(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	_, err := svc.ValidateAccessToken(ctx, "garbage-token")
	assert.Error(t, err)

	_, err = svc.ValidateRefreshToken(ctx, "garbage-token")
	assert.Error(t, err)
}

func TestJWTTokenService_RevokeNonexistentUser(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	// Revoking tokens for user with no tokens should not error
	err := svc.RevokeAllUserTokens(ctx, "nonexistent-user")
	assert.NoError(t, err)
}

func TestJWTTokenService_ValidateAccessToken_RevokedJTI(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	access, _, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	// Manually delete the JTI from Redis to simulate expiry
	claims, _ := svc.parseToken(access)
	jti := claims["jti"].(string)
	svc.redis.Del(ctx, jtiKey(jti))

	_, err = svc.ValidateAccessToken(ctx, access)
	assert.Error(t, err)
}

func TestJWTTokenService_GenerateWithCancelledContext(t *testing.T) {
	svc := setupTokenService(t)
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := svc.GenerateTokenPair(cancelCtx, "user-1", "test@example.com", "user")
	assert.Error(t, err)
}

func TestJWTTokenService_ExpiredToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)
	cfg := config.JWTConfig{
		Secret:     "test-secret-key-for-jwt",
		AccessTTL:  1 * time.Millisecond, // expires immediately
		RefreshTTL: 1 * time.Millisecond,
	}
	svc := NewJWTTokenService(containers.RedisClient, cfg).(*jwtTokenService)
	ctx := context.Background()

	access, refresh, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	_, err = svc.ValidateAccessToken(ctx, access)
	assert.Error(t, err)

	_, err = svc.ValidateRefreshToken(ctx, refresh)
	assert.Error(t, err)
}

func TestJWTTokenService_WrongSigningKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	containers := tc.SetupTestContainers(t)

	// Create token with one secret
	svc1 := NewJWTTokenService(containers.RedisClient, config.JWTConfig{
		Secret: "secret-1", AccessTTL: 15 * time.Minute, RefreshTTL: 168 * time.Hour,
	}).(*jwtTokenService)

	// Validate with different secret
	svc2 := NewJWTTokenService(containers.RedisClient, config.JWTConfig{
		Secret: "secret-2", AccessTTL: 15 * time.Minute, RefreshTTL: 168 * time.Hour,
	}).(*jwtTokenService)

	ctx := context.Background()
	access, _, err := svc1.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	_, err = svc2.ValidateAccessToken(ctx, access)
	assert.Error(t, err)
}

func TestJWTTokenService_AccessTokenAsRefresh_Rejected(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	access, _, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	// Using access token as refresh should fail
	_, err = svc.ValidateRefreshToken(ctx, access)
	assert.Error(t, err)
}

func TestJWTTokenService_RefreshTokenAsAccess_Rejected(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	_, refresh, err := svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	require.NoError(t, err)

	// Using refresh token as access should fail (type != "access")
	_, err = svc.ValidateAccessToken(ctx, refresh)
	assert.Error(t, err)
}

func TestJWTTokenService_RevokeWithCancelledContext(t *testing.T) {
	svc := setupTokenService(t)
	ctx := context.Background()

	_, _, err := svc.GenerateTokenPair(ctx, "user-cancel", "cancel@example.com", "user")
	require.NoError(t, err)

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	err = svc.RevokeAllUserTokens(cancelCtx, "user-cancel")
	assert.Error(t, err)
}
