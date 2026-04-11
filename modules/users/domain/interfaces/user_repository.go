package interfaces

import (
	"context"

	"booker/modules/users/domain/entities"
)

// UserRepository defines data access for users.
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, role string) (*entities.User, error)
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	List(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	Update(ctx context.Context, id, email, role, status string) (*entities.User, error)
}
