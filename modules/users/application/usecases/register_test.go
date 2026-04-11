package usecases

import (
	"context"
	"testing"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUseCase_Success(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRegisterUseCase(userSvc, tokenSvc)

	user := &entities.User{ID: "uuid-1", Email: "test@example.com", Role: "user"}
	userSvc.EXPECT().Create(mock.Anything, "test@example.com", "password123").Return(user, nil)
	tokenSvc.EXPECT().GenerateTokenPair(mock.Anything, "uuid-1", "test@example.com", "user").
		Return("access-token", "refresh-token", nil)

	result, err := uc.Execute(context.Background(), dto.RegisterDTO{
		Email: "test@example.com", Password: "password123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "uuid-1", result.User.ID)
	assert.Equal(t, "access-token", result.AccessToken)
	assert.Equal(t, "refresh-token", result.RefreshToken)
}

func TestRegisterUseCase_TokenGenerationFails(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRegisterUseCase(userSvc, tokenSvc)

	user := &entities.User{ID: "uuid-1", Email: "test@example.com", Role: "user"}
	userSvc.EXPECT().Create(mock.Anything, "test@example.com", "password123").Return(user, nil)
	tokenSvc.EXPECT().GenerateTokenPair(mock.Anything, "uuid-1", "test@example.com", "user").
		Return("", "", assert.AnError)

	result, err := uc.Execute(context.Background(), dto.RegisterDTO{
		Email: "test@example.com", Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRegisterUseCase_CreateFails(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRegisterUseCase(userSvc, tokenSvc)

	userSvc.EXPECT().Create(mock.Anything, "test@example.com", "password123").
		Return(nil, assert.AnError)

	result, err := uc.Execute(context.Background(), dto.RegisterDTO{
		Email: "test@example.com", Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}
