package token

import (
	"context"
	"fmt"
	"time"

	"booker/config"
	"booker/modules/users/domain"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type jwtTokenService struct {
	redis      *redis.Client
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTTokenService creates a Redis-backed JWT token service.
func NewJWTTokenService(redisClient *redis.Client, cfg config.JWTConfig) interfaces.TokenService {
	return &jwtTokenService{
		redis:      redisClient,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}
}

func (s *jwtTokenService) GenerateTokenPair(ctx context.Context, userID, email, role string) (string, string, error) {
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	// Access token
	accessClaims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"jti":   accessJTI,
		"exp":   time.Now().Add(s.accessTTL).Unix(),
		"iat":   time.Now().Unix(),
		"type":  "access",
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secret)
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		"sub":  userID,
		"jti":  refreshJTI,
		"exp":  time.Now().Add(s.refreshTTL).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.secret)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	// Whitelist both JTIs in Redis
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, jtiKey(accessJTI), userID, s.accessTTL)
	pipe.Set(ctx, jtiKey(refreshJTI), userID, s.refreshTTL)
	pipe.SAdd(ctx, userJTISetKey(userID), accessJTI, refreshJTI)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", fmt.Errorf("store JTIs in redis: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *jwtTokenService) ValidateAccessToken(ctx context.Context, tokenStr string) (*entities.AccessClaims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return nil, domain.ErrInvalidToken
	}

	jti, _ := claims["jti"].(string)
	if err := s.redis.Get(ctx, jtiKey(jti)).Err(); err != nil {
		return nil, domain.ErrInvalidToken
	}

	return &entities.AccessClaims{
		UserID: claims["sub"].(string),
		Email:  claims["email"].(string),
		Role:   claims["role"].(string),
		JTI:    jti,
	}, nil
}

func (s *jwtTokenService) ValidateRefreshToken(ctx context.Context, tokenStr string) (*entities.RefreshClaims, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, domain.ErrInvalidToken
	}

	jti, _ := claims["jti"].(string)

	// Single-use: delete on validation to prevent replay
	deleted, err := s.redis.Del(ctx, jtiKey(jti)).Result()
	if err != nil || deleted == 0 {
		return nil, domain.ErrInvalidToken
	}

	return &entities.RefreshClaims{
		UserID: claims["sub"].(string),
		JTI:    jti,
	}, nil
}

func (s *jwtTokenService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	setKey := userJTISetKey(userID)

	// Get all JTIs for this user
	jtis, err := s.redis.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}

	if len(jtis) == 0 {
		return nil
	}

	// Delete all JTI keys + the set itself
	pipe := s.redis.Pipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, jtiKey(jti))
	}
	pipe.Del(ctx, setKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *jwtTokenService) parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

func jtiKey(jti string) string           { return "jwt:jti:" + jti }
func userJTISetKey(userID string) string { return "jwt:user:" + userID }
