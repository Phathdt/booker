package interfaces

import (
	"context"

	"booker/modules/users/domain/entities"
)

// UserService defines business logic for users.
type UserService interface {
	Create(ctx context.Context, email, password string) (*entities.User, error)
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (*entities.User, error)
	List(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
}
