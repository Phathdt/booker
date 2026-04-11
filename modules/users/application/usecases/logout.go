package usecases

import (
	"context"

	"booker/modules/users/domain/interfaces"
)

type LogoutUseCase struct {
	tokenSvc interfaces.TokenService
}

func NewLogoutUseCase(tokenSvc interfaces.TokenService) *LogoutUseCase {
	return &LogoutUseCase{tokenSvc: tokenSvc}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, userID string) error {
	return uc.tokenSvc.RevokeAllUserTokens(ctx, userID)
}
