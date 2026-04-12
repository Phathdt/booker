package services

import (
	"context"
	"log/slog"

	"booker/modules/order/application/dto"
	"booker/modules/order/domain"
	"booker/modules/order/domain/entities"
	"booker/modules/order/domain/interfaces"

	"github.com/shopspring/decimal"
)

type orderService struct {
	repo         interfaces.OrderRepository
	walletClient interfaces.WalletClient
}

func NewOrderService(repo interfaces.OrderRepository, walletClient interfaces.WalletClient) interfaces.OrderService {
	return &orderService{repo: repo, walletClient: walletClient}
}

func (s *orderService) CreateOrder(
	ctx context.Context,
	userID string,
	req *dto.CreateOrderDTO,
) (*entities.Order, error) {
	if req.Price.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrInvalidPrice
	}
	if req.Quantity.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrInvalidQuantity
	}
	if req.Side != "buy" && req.Side != "sell" {
		return nil, domain.ErrInvalidSide
	}
	if req.Type != "limit" {
		return nil, domain.ErrInvalidOrderType
	}

	pair, err := s.repo.GetTradingPair(ctx, req.PairID)
	if err != nil {
		return nil, domain.ErrPairNotFound
	}
	if pair.Status != "active" {
		return nil, domain.ErrPairNotActive
	}
	if req.Quantity.LessThan(pair.MinQty) {
		return nil, domain.ErrBelowMinQty
	}
	if !req.Price.Mod(pair.TickSize).IsZero() {
		return nil, domain.ErrInvalidTickSize
	}

	// Determine hold asset and amount
	var holdAsset string
	var holdAmount decimal.Decimal
	if req.Side == "buy" {
		holdAsset = pair.QuoteAsset
		holdAmount = req.Price.Mul(req.Quantity)
	} else {
		holdAsset = pair.BaseAsset
		holdAmount = req.Quantity
	}

	if err := s.walletClient.HoldBalance(ctx, userID, holdAsset, holdAmount); err != nil {
		return nil, domain.ErrInsufficientBalance
	}

	order := &entities.Order{
		UserID:   userID,
		PairID:   req.PairID,
		Side:     req.Side,
		Type:     req.Type,
		Price:    req.Price,
		Quantity: req.Quantity,
	}

	created, err := s.repo.Create(ctx, order)
	if err != nil {
		// Rollback wallet hold
		releaseErr := s.walletClient.ReleaseBalance(ctx, userID, holdAsset, holdAmount)
		if releaseErr != nil {
			slog.ErrorContext(ctx, "CRITICAL: failed to release wallet hold after order creation failure",
				"user_id", userID,
				"asset_id", holdAsset,
				"amount", holdAmount.String(),
				"create_error", err.Error(),
				"release_error", releaseErr.Error(),
			)
		}
		return nil, err
	}

	return created, nil
}

func (s *orderService) CancelOrder(ctx context.Context, userID, orderID string) (*entities.Order, error) {
	order, err := s.repo.GetByIDAndUser(ctx, orderID, userID)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	if !order.IsCancellable() {
		return nil, domain.ErrOrderNotCancellable
	}

	pair, err := s.repo.GetTradingPair(ctx, order.PairID)
	if err != nil {
		return nil, domain.ErrPairNotFound
	}

	// Determine release asset and amount
	var releaseAsset string
	var releaseAmount decimal.Decimal
	remainingQty := order.RemainingQty()
	if order.Side == "buy" {
		releaseAsset = pair.QuoteAsset
		releaseAmount = order.Price.Mul(remainingQty)
	} else {
		releaseAsset = pair.BaseAsset
		releaseAmount = remainingQty
	}

	// Release wallet FIRST — if fails, order stays active (safe)
	if err := s.walletClient.ReleaseBalance(ctx, userID, releaseAsset, releaseAmount); err != nil {
		return nil, domain.ErrWalletServiceUnavailable.Wrap(err)
	}

	cancelled, err := s.repo.Cancel(ctx, orderID, userID)
	if err != nil {
		// Compensate: re-hold the released amount to prevent double-release
		if holdErr := s.walletClient.HoldBalance(ctx, userID, releaseAsset, releaseAmount); holdErr != nil {
			slog.ErrorContext(ctx, "CRITICAL: cancel failed after release, re-hold also failed",
				"user_id", userID,
				"order_id", orderID,
				"asset_id", releaseAsset,
				"amount", releaseAmount.String(),
				"cancel_error", err.Error(),
				"hold_error", holdErr.Error(),
			)
		}
		return nil, domain.ErrOrderNotCancellable
	}

	return cancelled, nil
}

func (s *orderService) GetOrder(ctx context.Context, userID, orderID string) (*entities.Order, error) {
	// Empty userID = inter-service call (gRPC), no user scoping
	if userID == "" {
		order, err := s.repo.GetByID(ctx, orderID)
		if err != nil {
			return nil, domain.ErrOrderNotFound
		}
		return order, nil
	}

	order, err := s.repo.GetByIDAndUser(ctx, orderID, userID)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) ListOrders(
	ctx context.Context,
	userID string,
	req *dto.ListOrdersDTO,
) ([]*entities.Order, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	var pairID, status *string
	if req.PairID != "" {
		pairID = &req.PairID
	}
	if req.Status != "" {
		status = &req.Status
	}

	return s.repo.List(ctx, userID, pairID, status, limit, req.Offset)
}

func (s *orderService) UpdateOrderFill(
	ctx context.Context,
	orderID string,
	filledQty decimal.Decimal,
	status string,
) (*entities.Order, error) {
	if filledQty.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrInvalidFillQty
	}

	// Validate status transition
	if status != "partial" && status != "filled" {
		return nil, domain.ErrOrderNotFillable
	}

	order, err := s.repo.UpdateFilledQty(ctx, orderID, filledQty, status)
	if err != nil {
		return nil, domain.ErrOrderNotFillable
	}

	return order, nil
}
