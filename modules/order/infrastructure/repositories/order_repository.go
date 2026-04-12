package repositories

import (
	"context"
	"errors"

	"booker/modules/order/domain"
	"booker/modules/order/domain/entities"
	"booker/modules/order/domain/interfaces"
	"booker/modules/order/infrastructure/gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type orderRepository struct {
	q *gen.Queries
}

func NewOrderRepository(pool *pgxpool.Pool) interfaces.OrderRepository {
	return &orderRepository{q: gen.New(pool)}
}

func (r *orderRepository) GetTradingPair(ctx context.Context, pairID string) (*interfaces.TradingPair, error) {
	row, err := r.q.GetTradingPair(ctx, pairID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPairNotFound
		}
		return nil, err
	}
	return &interfaces.TradingPair{
		ID:         row.ID,
		BaseAsset:  row.BaseAsset,
		QuoteAsset: row.QuoteAsset,
		Status:     row.Status,
		MinQty:     row.MinQty,
		TickSize:   row.TickSize,
	}, nil
}

func (r *orderRepository) Create(ctx context.Context, order *entities.Order) (*entities.Order, error) {
	row, err := r.q.CreateOrder(ctx, gen.CreateOrderParams{
		UserID:   order.UserID,
		PairID:   order.PairID,
		Side:     order.Side,
		Type:     order.Type,
		Price:    order.Price,
		Quantity: order.Quantity,
	})
	if err != nil {
		return nil, err
	}
	return toEntity(row), nil
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*entities.Order, error) {
	row, err := r.q.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *orderRepository) GetByIDAndUser(ctx context.Context, id, userID string) (*entities.Order, error) {
	row, err := r.q.GetOrderByIDAndUser(ctx, gen.GetOrderByIDAndUserParams{
		ID: id, UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *orderRepository) List(
	ctx context.Context,
	userID string,
	pairID, status *string,
	limit, offset int32,
) ([]*entities.Order, error) {
	rows, err := r.q.ListOrders(ctx, gen.ListOrdersParams{
		UserID: userID,
		PairID: pairID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	orders := make([]*entities.Order, len(rows))
	for i, row := range rows {
		orders[i] = toEntity(row)
	}
	return orders, nil
}

func (r *orderRepository) Cancel(ctx context.Context, id, userID string) (*entities.Order, error) {
	row, err := r.q.CancelOrder(ctx, gen.CancelOrderParams{
		ID: id, UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotCancellable
		}
		return nil, err
	}
	return toEntity(row), nil
}

func (r *orderRepository) UpdateFilledQty(
	ctx context.Context,
	id string,
	filledQty decimal.Decimal,
	status string,
) (*entities.Order, error) {
	row, err := r.q.UpdateOrderFilledQty(ctx, gen.UpdateOrderFilledQtyParams{
		ID: id, FilledQty: filledQty, Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFillable
		}
		return nil, err
	}
	return toEntity(row), nil
}

func toEntity(row gen.Order) *entities.Order {
	return &entities.Order{
		ID:        row.ID,
		UserID:    row.UserID,
		PairID:    row.PairID,
		Side:      row.Side,
		Type:      row.Type,
		Price:     row.Price,
		Quantity:  row.Quantity,
		FilledQty: row.FilledQty,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
