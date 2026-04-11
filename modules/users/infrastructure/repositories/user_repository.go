package repositories

import (
	"context"

	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"
	"booker/modules/users/infrastructure/gen"

	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	q *gen.Queries
}

// NewUserRepository creates a new UserRepository backed by SQLC.
func NewUserRepository(pool *pgxpool.Pool) interfaces.UserRepository {
	return &userRepository{q: gen.New(pool)}
}

func (r *userRepository) Create(ctx context.Context, email, passwordHash, role string) (*entities.User, error) {
	row, err := r.q.CreateUser(ctx, gen.CreateUserParams{
		Email:    email,
		Password: passwordHash,
		Role:     role,
	})
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	rows, err := r.q.ListUsers(ctx, gen.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := r.q.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	users := make([]*entities.User, len(rows))
	for i, row := range rows {
		users[i] = toEntity(row)
	}
	return users, count, nil
}

func (r *userRepository) Update(ctx context.Context, id, email, role, status string) (*entities.User, error) {
	row, err := r.q.UpdateUser(ctx, gen.UpdateUserParams{
		ID:     id,
		Email:  email,
		Role:   role,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func toEntity(row gen.User) *entities.User {
	return &entities.User{
		ID:        row.ID,
		Email:     row.Email,
		Password:  row.Password,
		Role:      row.Role,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
