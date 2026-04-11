package usecases

import (
	"context"
	"testing"

	"booker/modules/users/domain/interfaces/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutUseCase_Success(t *testing.T) {
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewLogoutUseCase(tokenSvc)

	tokenSvc.EXPECT().RevokeAllUserTokens(mock.Anything, "uuid-1").Return(nil)

	err := uc.Execute(context.Background(), "uuid-1")
	assert.NoError(t, err)
}

func TestLogoutUseCase_Error(t *testing.T) {
	tokenSvc := mocks.NewMockTokenService(t)
	uc := NewLogoutUseCase(tokenSvc)

	tokenSvc.EXPECT().RevokeAllUserTokens(mock.Anything, "uuid-1").Return(assert.AnError)

	err := uc.Execute(context.Background(), "uuid-1")
	assert.Error(t, err)
}
