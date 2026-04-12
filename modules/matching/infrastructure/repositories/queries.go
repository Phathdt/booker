package repositories

import (
	"context"

	"booker/modules/matching/infrastructure/gen"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Queries wraps SQLC generated queries for direct use in CLI wiring.
type Queries struct {
	q *gen.Queries
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{q: gen.New(pool)}
}

func (q *Queries) ListActiveTradingPairs(ctx context.Context) ([]gen.TradingPair, error) {
	return q.q.ListActiveTradingPairs(ctx)
}

func (q *Queries) ListOpenOrdersByPair(ctx context.Context, pairID string) ([]gen.Order, error) {
	return q.q.ListOpenOrdersByPair(ctx, pairID)
}
