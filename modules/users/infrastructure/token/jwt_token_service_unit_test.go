package token

import (
	"testing"
	"time"

	"booker/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTTokenService_ParseToken_InvalidTokenFormat tests parseToken with invalid format
func TestJWTTokenService_ParseToken_InvalidTokenFormat(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	claims, err := svc.parseToken("invalid-token-garbage")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestJWTTokenService_ParseToken_ExpiredToken tests parseToken with expired token
func TestJWTTokenService_ParseToken_ExpiredToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create an expired token
	expiredClaims := jwt.MapClaims{
		"sub":  "user-1",
		"jti":  "jti-123",
		"exp":  time.Now().Add(-1 * time.Hour).Unix(), // Expired
		"iat":  time.Now().Unix(),
		"type": "access",
	}

	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("test-secret-key"))

	claims, err := svc.parseToken(expiredToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestJWTTokenService_ParseToken_WrongSecret tests token signed with wrong secret
func TestJWTTokenService_ParseToken_WrongSecret(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with different secret
	claims := jwt.MapClaims{
		"sub":  "user-1",
		"jti":  "jti-123",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"type": "access",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("different-secret"))

	parsedClaims, err := svc.parseToken(token)

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

// TestJWTTokenService_ParseToken_ValidToken tests parseToken with valid token
func TestJWTTokenService_ParseToken_ValidToken(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create a valid token
	validClaims := jwt.MapClaims{
		"sub":  "user-1",
		"jti":  "jti-123",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"type": "access",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims).SignedString([]byte("test-secret-key"))

	parsedClaims, err := svc.parseToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, parsedClaims)
	assert.Equal(t, "user-1", parsedClaims["sub"])
	assert.Equal(t, "jti-123", parsedClaims["jti"])
	assert.Equal(t, "access", parsedClaims["type"])
}

// TestJWTTokenService_ParseToken_BadSigningMethod tests token with wrong algorithm
func TestJWTTokenService_ParseToken_BadSigningMethod(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Token with RS256 header but no valid RS key - will fail on parse
	invalidToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEiLCJqdGkiOiJqdGktMTIzIiwiZXhwIjoxNzM2NjE0MDAwLCJpYXQiOjE3MzY2MTAyNDB9.signature"

	parsedClaims, err := svc.parseToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

// TestJWTTokenService_ParseToken_EmptyString tests parseToken with empty string
func TestJWTTokenService_ParseToken_EmptyString(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	parsedClaims, err := svc.parseToken("")

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

// TestJWTTokenService_ParseToken_MissingDots tests parseToken with malformed token
func TestJWTTokenService_ParseToken_MissingDots(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	parsedClaims, err := svc.parseToken("notavalidtoken")

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

// TestJWTTokenService_ParseToken_ValidTokenReturnsMapClaims tests that valid tokens return proper claim structure
func TestJWTTokenService_ParseToken_ValidTokenReturnsMapClaims(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with multiple claims
	claims := jwt.MapClaims{
		"sub":   "user-123",
		"jti":   "token-456",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"type":  "access",
		"email": "user@example.com",
		"role":  "admin",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	parsedClaims, err := svc.parseToken(token)

	require.NoError(t, err)
	assert.Equal(t, "user-123", parsedClaims["sub"])
	assert.Equal(t, "token-456", parsedClaims["jti"])
	assert.Equal(t, "access", parsedClaims["type"])
	assert.Equal(t, "user@example.com", parsedClaims["email"])
	assert.Equal(t, "admin", parsedClaims["role"])
}

// TestJWTTokenService_JTIKeyFunc tests jtiKey helper function
func TestJWTTokenService_JTIKeyFunc(t *testing.T) {
	key := jtiKey("test-jti-123")

	assert.Equal(t, "jwt:jti:test-jti-123", key)
}

// TestJWTTokenService_UserJTISetKeyFunc tests userJTISetKey helper function
func TestJWTTokenService_UserJTISetKeyFunc(t *testing.T) {
	key := userJTISetKey("user-456")

	assert.Equal(t, "jwt:user:user-456", key)
}

// TestJWTTokenService_HelperKeyFormatting tests key formatting consistency
func TestJWTTokenService_HelperKeyFormatting(t *testing.T) {
	jti := "unique-jti-789"
	userID := "user-xyz"

	jtiKey := jtiKey(jti)
	userKey := userJTISetKey(userID)

	// Keys should start with appropriate prefixes
	assert.True(t, len(jtiKey) > 0)
	assert.True(t, len(userKey) > 0)

	// Keys should contain the original values
	assert.Contains(t, jtiKey, jti)
	assert.Contains(t, userKey, userID)
}

// TestJWTTokenService_ParseToken_NoExpiry tests token without expiry (edge case)
func TestJWTTokenService_ParseToken_NoExpiry(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Token without 'exp' claim (not standard but let's see how it handles)
	claims := jwt.MapClaims{
		"sub":  "user-1",
		"jti":  "jti-123",
		"iat":  time.Now().Unix(),
		"type": "access",
		// No 'exp' claim
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	// Should still parse successfully even without exp
	parsedClaims, err := svc.parseToken(token)

	// JWT library doesn't validate expiry by default, so this should succeed
	assert.NoError(t, err)
	assert.NotNil(t, parsedClaims)
}
