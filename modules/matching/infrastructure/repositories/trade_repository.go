package repositories

import (
	"context"
	"errors"

	"booker/modules/matching/domain"
	"booker/modules/matching/domain/entities"
	"booker/modules/matching/domain/interfaces"
	"booker/modules/matching/infrastructure/gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tradeRepository struct {
	q *gen.Queries
}

func NewTradeRepository(pool *pgxpool.Pool) interfaces.TradeRepository {
	return &tradeRepository{q: gen.New(pool)}
}

func (r *tradeRepository) Create(ctx context.Context, trade *entities.Trade) (*entities.Trade, error) {
	row, err := r.q.CreateTrade(ctx, gen.CreateTradeParams{
		PairID:      trade.PairID,
		BuyOrderID:  trade.BuyOrderID,
		SellOrderID: trade.SellOrderID,
		Price:       trade.Price,
		Quantity:    trade.Quantity,
		BuyerID:     trade.BuyerID,
		SellerID:    trade.SellerID,
	})
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func (r *tradeRepository) GetByID(ctx context.Context, id string) (*entities.Trade, error) {
	row, err := r.q.GetTradeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTradeNotFound
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *tradeRepository) ListByPair(
	ctx context.Context,
	pairID string,
	limit, offset int32,
) ([]*entities.Trade, error) {
	rows, err := r.q.ListTradesByPair(ctx, gen.ListTradesByPairParams{
		PairID: pairID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	trades := make([]*entities.Trade, len(rows))
	for i, row := range rows {
		trades[i] = toEntity(row)
	}
	return trades, nil
}

func toEntity(row gen.Trade) *entities.Trade {
	return &entities.Trade{
		ID:          row.ID,
		PairID:      row.PairID,
		BuyOrderID:  row.BuyOrderID,
		SellOrderID: row.SellOrderID,
		Price:       row.Price,
		Quantity:    row.Quantity,
		BuyerID:     row.BuyerID,
		SellerID:    row.SellerID,
		ExecutedAt:  row.ExecutedAt,
	}
}
