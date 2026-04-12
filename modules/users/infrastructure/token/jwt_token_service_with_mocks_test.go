package token

import (
	"context"
	"fmt"
	"testing"
	"time"

	"booker/config"
	"booker/modules/users/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestNewJWTTokenService tests constructor
func TestNewJWTTokenService(t *testing.T) {
	mockRedis := &redis.Client{}
	cfg := config.JWTConfig{
		Secret:     "test-secret",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := NewJWTTokenService(mockRedis, cfg)

	assert.NotNil(t, svc)
	ts := svc.(*jwtTokenService)
	assert.Equal(t, mockRedis, ts.redis)
	assert.Equal(t, []byte("test-secret"), ts.secret)
	assert.Equal(t, 15*time.Minute, ts.accessTTL)
	assert.Equal(t, 7*24*time.Hour, ts.refreshTTL)
}

// TestGenerateTokenPair_RedisFailure tests handling of Redis pipeline failure
func TestGenerateTokenPair_RedisFailure(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil, // Will cause panic if accessed
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	ctx := context.Background()

	// This should panic because redis is nil
	assert.Panics(t, func() {
		svc.GenerateTokenPair(ctx, "user-1", "test@example.com", "user")
	})
}

// TestValidateAccessToken_MissingJTI tests handling of token without JTI
func TestValidateAccessToken_MissingJTI(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token without JTI
	claims := jwt.MapClaims{
		"sub":   "user-1",
		"email": "test@example.com",
		"role":  "user",
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
		"type":  "access",
		// No "jti" claim
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateAccessToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestValidateAccessToken_MissingType tests handling of token without type
func TestValidateAccessToken_MissingType(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token without type
	claims := jwt.MapClaims{
		"sub":   "user-1",
		"email": "test@example.com",
		"role":  "user",
		"jti":   "jti-123",
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
		// No "type" claim
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateAccessToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestValidateAccessToken_WrongType tests validation of refresh token as access
func TestValidateAccessToken_WrongType(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with type "refresh"
	claims := jwt.MapClaims{
		"sub":   "user-1",
		"email": "test@example.com",
		"role":  "user",
		"jti":   "jti-123",
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
		"type":  "refresh",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateAccessToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestValidateRefreshToken_MissingType tests handling of refresh token without type
func TestValidateRefreshToken_MissingType(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token without type
	claims := jwt.MapClaims{
		"sub": "user-1",
		"jti": "jti-123",
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		// No "type" claim
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateRefreshToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestValidateRefreshToken_WrongType tests validation of access token as refresh
func TestValidateRefreshToken_WrongType(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with type "access"
	claims := jwt.MapClaims{
		"sub":  "user-1",
		"jti":  "jti-123",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"type": "access",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateRefreshToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestValidateRefreshToken_MissingJTI tests handling of refresh token without JTI
func TestValidateRefreshToken_MissingJTI(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token without JTI
	claims := jwt.MapClaims{
		"sub":  "user-1",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
		// No "jti" claim
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateRefreshToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestParseToken_WrongAlgorithmMethod tests that wrong algorithm is rejected
func TestParseToken_WrongAlgorithmMethod(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create a token with "none" algorithm (invalid)
	// JWT with alg: none should be rejected
	invalidToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEifQ."

	_, err := svc.parseToken(invalidToken)
	assert.Error(t, err)
}

// TestValidateAccessToken_EmptyJTI tests handling of empty JTI
func TestValidateAccessToken_EmptyJTI(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with empty JTI
	claims := jwt.MapClaims{
		"sub":   "user-1",
		"email": "test@example.com",
		"role":  "user",
		"jti":   "", // Empty JTI
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
		"type":  "access",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateAccessToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestValidateRefreshToken_EmptyJTI tests handling of empty JTI in refresh token
func TestValidateRefreshToken_EmptyJTI(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with empty JTI
	claims := jwt.MapClaims{
		"sub":  "user-1",
		"jti":  "", // Empty JTI
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	_, err := svc.ValidateRefreshToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// TestParseToken_InvalidJWTStructure tests parsing of completely invalid JWT
func TestParseToken_InvalidJWTStructure(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	invalidTokens := []string{
		"",
		"single-part",
		"two.parts",
		"a.b.c.d",
		"notajwt.atall.token",
		"eyJ.eyJ.eyJ", // Invalid base64
	}

	for _, invalidToken := range invalidTokens {
		_, err := svc.parseToken(invalidToken)
		assert.Error(t, err, "should error for token: "+invalidToken)
		assert.Equal(t, domain.ErrInvalidToken, err)
	}
}

// TestNewJWTTokenService_WithDifferentTTLs tests constructor with various TTL configurations
func TestNewJWTTokenService_WithDifferentTTLs(t *testing.T) {
	testCases := []struct {
		name       string
		accessTTL  time.Duration
		refreshTTL time.Duration
	}{
		{"1 hour access, 30 days refresh", time.Hour, 30 * 24 * time.Hour},
		{"5 min access, 1 year refresh", 5 * time.Minute, 365 * 24 * time.Hour},
		{"30 min access, 1 week refresh", 30 * time.Minute, 7 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRedis := &redis.Client{}
			cfg := config.JWTConfig{
				Secret:     "test-secret",
				AccessTTL:  tc.accessTTL,
				RefreshTTL: tc.refreshTTL,
			}

			svc := NewJWTTokenService(mockRedis, cfg)
			ts := svc.(*jwtTokenService)

			assert.Equal(t, tc.accessTTL, ts.accessTTL)
			assert.Equal(t, tc.refreshTTL, ts.refreshTTL)
		})
	}
}

// TestNewJWTTokenService_WithVariousSecrets tests constructor with various secret configurations
func TestNewJWTTokenService_WithVariousSecrets(t *testing.T) {
	secrets := []string{
		"simple-secret",
		"very-long-secret-with-many-characters-and-numbers-12345",
		"secret-with-special-chars-!@#$%^&*()",
		"unicode-secret-密码-🔐",
	}

	for _, secret := range secrets {
		t.Run(fmt.Sprintf("secret_%d_chars", len(secret)), func(t *testing.T) {
			mockRedis := &redis.Client{}
			cfg := config.JWTConfig{
				Secret:     secret,
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 7 * 24 * time.Hour,
			}

			svc := NewJWTTokenService(mockRedis, cfg)
			ts := svc.(*jwtTokenService)

			assert.Equal(t, []byte(secret), ts.secret)
		})
	}
}

// TestParseToken_WrongSigningMethodReject tests that non-HMAC algorithms are rejected
func TestParseToken_WrongSigningMethodReject(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Create token with different signing method (will fail signature validation)
	// We'll use a real token with RS256 which can't be verified with HS256
	rsaToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEiLCJqdGkiOiJqdGktMTIzIiwiZXhwIjo5OTk5OTk5OTk5LCJpYXQiOjE2NzA0NjU2MDAsInR5cGUiOiJhY2Nlc3MifQ.invalid-signature-here"

	_, err := svc.parseToken(rsaToken)
	assert.Error(t, err)
}

// TestParseToken_RobustToInvalidClaims tests that parseToken handles invalid claim types
func TestParseToken_RobustToInvalidClaims(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	svc := &jwtTokenService{
		redis:      nil,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	// Token with mixed types should still parse successfully
	claims := jwt.MapClaims{
		"sub":      "user-1",
		"email":    "test@example.com",
		"role":     "user",
		"jti":      "jti-123",
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
		"number":   42,
		"float":    3.14,
		"boolean":  true,
		"null_val": nil,
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-key"))

	parsed, err := svc.parseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)

	// Verify various claims are present
	assert.Equal(t, "user-1", parsed["sub"])
	assert.Equal(t, float64(42), parsed["number"])
	assert.Equal(t, true, parsed["boolean"])
}
