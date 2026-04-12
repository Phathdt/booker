package services

import (
	"context"
	"log/slog"
	"time"

	"booker/modules/order/application/dto"
	"booker/modules/order/domain"
	"booker/modules/order/domain/entities"
	"booker/modules/order/domain/interfaces"
	pkgnats "booker/pkg/nats"

	"github.com/shopspring/decimal"
)

type orderService struct {
	repo           interfaces.OrderRepository
	walletClient   interfaces.WalletClient
	matchingClient interfaces.MatchingClient
	orderPublisher pkgnats.OrderPublisher
}

func NewOrderService(
	repo interfaces.OrderRepository,
	walletClient interfaces.WalletClient,
	matchingClient interfaces.MatchingClient,
	orderPublisher pkgnats.OrderPublisher,
) interfaces.OrderService {
	return &orderService{
		repo:           repo,
		walletClient:   walletClient,
		matchingClient: matchingClient,
		orderPublisher: orderPublisher,
	}
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
		return nil, err
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

	// Submit to matching engine (fire-and-forget for REST response)
	if s.matchingClient != nil {
		if err := s.matchingClient.SubmitOrder(ctx, created); err != nil {
			slog.WarnContext(ctx, "failed to submit order to matching engine",
				"order_id", created.ID, "error", err.Error())
		}
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

	// Remove from matching engine book (best-effort)
	if s.matchingClient != nil {
		if err := s.matchingClient.CancelOrder(ctx, order.PairID, orderID); err != nil {
			slog.WarnContext(ctx, "failed to cancel order in matching engine",
				"order_id", orderID, "error", err.Error())
		}
	}

	// Release wallet — if fails, order stays active (safe)
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

	s.publishOrderEvent(ctx, cancelled)
	return cancelled, nil
}

func (s *orderService) GetOrder(ctx context.Context, userID, orderID string) (*entities.Order, error) {
	if userID == "" {
		return nil, domain.ErrOrderNotFound
	}

	order, err := s.repo.GetByIDAndUser(ctx, orderID, userID)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

// GetOrderInternal retrieves an order without user scoping, for inter-service use only.
func (s *orderService) GetOrderInternal(ctx context.Context, orderID string) (*entities.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
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

	// Fetch current order to validate fill quantity is monotonically increasing
	current, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	if filledQty.LessThan(current.FilledQty) {
		return nil, domain.ErrFillQtyBackward
	}

	order, err := s.repo.UpdateFilledQty(ctx, orderID, filledQty, status)
	if err != nil {
		return nil, domain.ErrOrderNotFillable
	}

	s.publishOrderEvent(ctx, order)
	return order, nil
}

func (s *orderService) publishOrderEvent(ctx context.Context, order *entities.Order) {
	if s.orderPublisher == nil {
		return
	}
	event := &pkgnats.OrderEvent{
		OrderID:   order.ID,
		UserID:    order.UserID,
		PairID:    order.PairID,
		Side:      order.Side,
		Price:     order.Price.String(),
		Quantity:  order.Quantity.String(),
		FilledQty: order.FilledQty.String(),
		Status:    order.Status,
		UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
	}
	if err := s.orderPublisher.PublishOrderUpdate(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to publish order event",
			"order_id", order.ID, "error", err.Error())
	}
}
