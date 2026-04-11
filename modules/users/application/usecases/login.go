package usecases

import (
	"context"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"
)

// LoginResult holds the login output.
type LoginResult struct {
	User         *entities.User
	AccessToken  string
	RefreshToken string
}

type LoginUseCase struct {
	userSvc  interfaces.UserService
	tokenSvc interfaces.TokenService
}

func NewLoginUseCase(userSvc interfaces.UserService, tokenSvc interfaces.TokenService) *LoginUseCase {
	return &LoginUseCase{userSvc: userSvc, tokenSvc: tokenSvc}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input dto.LoginDTO) (*LoginResult, error) {
	user, err := uc.userSvc.ValidateCredentials(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	access, refresh, err := uc.tokenSvc.GenerateTokenPair(ctx, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{User: user, AccessToken: access, RefreshToken: refresh}, nil
}
