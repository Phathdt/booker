package usecases

import (
	"context"

	"booker/modules/users/domain/interfaces"
)

// RefreshResult holds the refresh token output.
type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type RefreshTokenUseCase struct {
	userSvc  interfaces.UserService
	tokenSvc interfaces.TokenService
}

func NewRefreshTokenUseCase(userSvc interfaces.UserService, tokenSvc interfaces.TokenService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{userSvc: userSvc, tokenSvc: tokenSvc}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	claims, err := uc.tokenSvc.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user to ensure still active and get latest data
	user, err := uc.userSvc.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	access, refresh, err := uc.tokenSvc.GenerateTokenPair(ctx, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{AccessToken: access, RefreshToken: refresh}, nil
}
