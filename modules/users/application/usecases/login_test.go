package usecases

import (
	"context"
	"testing"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginUseCase_Success(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewLoginUseCase(userSvc, tokenSvc)

	user := &entities.User{ID: "uuid-1", Email: "test@example.com", Role: "user"}
	userSvc.EXPECT().ValidateCredentials(mock.Anything, "test@example.com", "password123").Return(user, nil)
	tokenSvc.EXPECT().GenerateTokenPair(mock.Anything, "uuid-1", "test@example.com", "user").
		Return("access-token", "refresh-token", nil)

	result, err := uc.Execute(context.Background(), dto.LoginDTO{
		Email: "test@example.com", Password: "password123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "access-token", result.AccessToken)
}

func TestLoginUseCase_TokenGenerationFails(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewLoginUseCase(userSvc, tokenSvc)

	user := &entities.User{ID: "uuid-1", Email: "test@example.com", Role: "user"}
	userSvc.EXPECT().ValidateCredentials(mock.Anything, "test@example.com", "password123").Return(user, nil)
	tokenSvc.EXPECT().GenerateTokenPair(mock.Anything, "uuid-1", "test@example.com", "user").
		Return("", "", assert.AnError)

	result, err := uc.Execute(context.Background(), dto.LoginDTO{
		Email: "test@example.com", Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLoginUseCase_InvalidCredentials(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewLoginUseCase(userSvc, tokenSvc)

	userSvc.EXPECT().ValidateCredentials(mock.Anything, "test@example.com", "wrong").
		Return(nil, domain.ErrInvalidCredentials)

	result, err := uc.Execute(context.Background(), dto.LoginDTO{
		Email: "test@example.com", Password: "wrong",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
}
