package interfaces

import (
	"context"

	"booker/modules/users/domain/entities"
)

// TokenService defines JWT token operations.
type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID, email, role string) (accessToken, refreshToken string, err error)
	ValidateAccessToken(ctx context.Context, token string) (*entities.AccessClaims, error)
	ValidateRefreshToken(ctx context.Context, token string) (*entities.RefreshClaims, error)
	RevokeAllUserTokens(ctx context.Context, userID string) error
}
