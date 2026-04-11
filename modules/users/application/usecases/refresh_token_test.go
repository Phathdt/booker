package usecases

import (
	"context"
	"testing"

	"booker/modules/users/domain"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshTokenUseCase_Success(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRefreshTokenUseCase(userSvc, tokenSvc)

	tokenSvc.EXPECT().ValidateRefreshToken(mock.Anything, "old-refresh").
		Return(&entities.RefreshClaims{UserID: "uuid-1", JTI: "jti-1"}, nil)
	userSvc.EXPECT().GetByID(mock.Anything, "uuid-1").
		Return(&entities.User{ID: "uuid-1", Email: "test@example.com", Role: "user"}, nil)
	tokenSvc.EXPECT().GenerateTokenPair(mock.Anything, "uuid-1", "test@example.com", "user").
		Return("new-access", "new-refresh", nil)

	result, err := uc.Execute(context.Background(), "old-refresh")

	assert.NoError(t, err)
	assert.Equal(t, "new-access", result.AccessToken)
	assert.Equal(t, "new-refresh", result.RefreshToken)
}

func TestRefreshTokenUseCase_InvalidToken(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRefreshTokenUseCase(userSvc, tokenSvc)

	tokenSvc.EXPECT().ValidateRefreshToken(mock.Anything, "bad-token").
		Return(nil, domain.ErrInvalidToken)

	result, err := uc.Execute(context.Background(), "bad-token")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRefreshTokenUseCase_TokenGenerationFails(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRefreshTokenUseCase(userSvc, tokenSvc)

	tokenSvc.EXPECT().ValidateRefreshToken(mock.Anything, "old-refresh").
		Return(&entities.RefreshClaims{UserID: "uuid-1", JTI: "jti-1"}, nil)
	userSvc.EXPECT().GetByID(mock.Anything, "uuid-1").
		Return(&entities.User{ID: "uuid-1", Email: "test@example.com", Role: "user"}, nil)
	tokenSvc.EXPECT().GenerateTokenPair(mock.Anything, "uuid-1", "test@example.com", "user").
		Return("", "", assert.AnError)

	result, err := uc.Execute(context.Background(), "old-refresh")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRefreshTokenUseCase_UserNotFound(t *testing.T) {
	userSvc := mocks.NewMockUserService(t)
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewRefreshTokenUseCase(userSvc, tokenSvc)

	tokenSvc.EXPECT().ValidateRefreshToken(mock.Anything, "old-refresh").
		Return(&entities.RefreshClaims{UserID: "uuid-999", JTI: "jti-1"}, nil)
	userSvc.EXPECT().GetByID(mock.Anything, "uuid-999").
		Return(nil, domain.ErrUserNotFound)

	result, err := uc.Execute(context.Background(), "old-refresh")

	assert.Error(t, err)
	assert.Nil(t, result)
}
