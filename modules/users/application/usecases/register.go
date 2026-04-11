package usecases

import (
	"context"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"
)

// RegisterResult holds the registration output.
type RegisterResult struct {
	User         *entities.User
	AccessToken  string
	RefreshToken string
}

type RegisterUseCase struct {
	userSvc  interfaces.UserService
	tokenSvc interfaces.TokenService
}

func NewRegisterUseCase(userSvc interfaces.UserService, tokenSvc interfaces.TokenService) *RegisterUseCase {
	return &RegisterUseCase{userSvc: userSvc, tokenSvc: tokenSvc}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input dto.RegisterDTO) (*RegisterResult, error) {
	user, err := uc.userSvc.Create(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	access, refresh, err := uc.tokenSvc.GenerateTokenPair(ctx, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &RegisterResult{User: user, AccessToken: access, RefreshToken: refresh}, nil
}
