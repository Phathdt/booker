package services

import (
	"context"
	"fmt"
	"testing"

	"booker/modules/users/domain"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Create_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(nil, domain.ErrUserNotFound)
	repo.EXPECT().Create(mock.Anything, "test@example.com", mock.AnythingOfType("string"), "user").
		Return(&entities.User{
			ID:    "uuid-1",
			Email: "test@example.com",
			Role:  "user",
		}, nil)

	user, err := svc.Create(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "uuid-1", user.ID)
}

func TestUserService_Create_HashError(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	// Password too long for bcrypt (>72 bytes triggers error)
	longPassword := string(make([]byte, 100))
	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(nil, fmt.Errorf("not found"))

	user, err := svc.Create(context.Background(), "test@example.com", longPassword)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_Create_EmailExists(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&entities.User{
		ID:    "uuid-1",
		Email: "test@example.com",
	}, nil)

	user, err := svc.Create(context.Background(), "test@example.com", "password123")

	assert.Nil(t, user)
	assert.Equal(t, domain.ErrEmailAlreadyExists, err)
}

func TestUserService_ValidateCredentials_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	// bcrypt hash of "password123"
	hashed := hashPassword(t, "password123")
	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&entities.User{
		ID:       "uuid-1",
		Email:    "test@example.com",
		Password: hashed,
		Status:   entities.StatusActive,
	}, nil)

	user, err := svc.ValidateCredentials(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "uuid-1", user.ID)
}

func TestUserService_ValidateCredentials_UserNotFound(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().GetByEmail(mock.Anything, "nonexistent@example.com").Return(nil, fmt.Errorf("not found"))

	user, err := svc.ValidateCredentials(context.Background(), "nonexistent@example.com", "password123")
	assert.Nil(t, user)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
}

func TestUserService_ValidateCredentials_WrongPassword(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	hashed := hashPassword(t, "correct-password")
	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&entities.User{
		ID:       "uuid-1",
		Email:    "test@example.com",
		Password: hashed,
		Status:   entities.StatusActive,
	}, nil)

	user, err := svc.ValidateCredentials(context.Background(), "test@example.com", "wrong-password")

	assert.Nil(t, user)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
}

func TestUserService_ValidateCredentials_Inactive(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	hashed := hashPassword(t, "password123")
	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&entities.User{
		ID:       "uuid-1",
		Email:    "test@example.com",
		Password: hashed,
		Status:   entities.StatusBanned,
	}, nil)

	user, err := svc.ValidateCredentials(context.Background(), "test@example.com", "password123")

	assert.Nil(t, user)
	assert.Equal(t, domain.ErrUserInactive, err)
}

func TestUserService_GetByID_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().GetByID(mock.Anything, "uuid-1").Return(&entities.User{
		ID: "uuid-1", Email: "test@example.com",
	}, nil)

	user, err := svc.GetByID(context.Background(), "uuid-1")
	assert.NoError(t, err)
	assert.Equal(t, "uuid-1", user.ID)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().GetByID(mock.Anything, "uuid-999").Return(nil, fmt.Errorf("no rows"))

	user, err := svc.GetByID(context.Background(), "uuid-999")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_GetByEmail_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&entities.User{
		ID: "uuid-1", Email: "test@example.com",
	}, nil)

	user, err := svc.GetByEmail(context.Background(), "  Test@Example.COM  ")
	assert.NoError(t, err)
	assert.Equal(t, "uuid-1", user.ID)
}

func TestUserService_List_Success(t *testing.T) {
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.EXPECT().List(mock.Anything, 10, 0).Return([]*entities.User{
		{ID: "uuid-1"}, {ID: "uuid-2"},
	}, int64(2), nil)

	users, total, err := svc.List(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}
