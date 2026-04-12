package services

import (
	"context"
	"errors"
	"strings"

	"booker/modules/users/domain"
	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo interfaces.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo interfaces.UserRepository) interfaces.UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, email, password string) (*entities.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, email, string(hash), entities.RoleUser)
}

func (s *userService) GetByID(ctx context.Context, id string) (*entities.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound.Wrap(err)
	}
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return s.repo.GetByEmail(ctx, email)
}

func (s *userService) ValidateCredentials(ctx context.Context, email, password string) (*entities.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status != entities.StatusActive {
		return nil, domain.ErrUserInactive
	}

	return user, nil
}

func (s *userService) List(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.repo.List(ctx, limit, offset)
}
