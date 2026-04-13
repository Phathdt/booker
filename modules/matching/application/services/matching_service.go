package services

import (
	"context"
	"log/slog"
	"time"

	"booker/modules/matching/domain"
	"booker/modules/matching/domain/entities"
	"booker/modules/matching/domain/interfaces"
	"booker/modules/matching/engine"
	pkgnats "booker/pkg/nats"

	"github.com/shopspring/decimal"
)

// PairInfo holds cached trading pair metadata for settlement.
type PairInfo struct {
	ID         string
	BaseAsset  string
	QuoteAsset string
}

type matchingService struct {
	engines            map[string]*engine.Engine
	tradeRepo          interfaces.TradeRepository
	orderClient        interfaces.OrderClient
	walletClient       interfaces.WalletClient
	tradePublisher     pkgnats.TradePublisher
	orderBookPublisher pkgnats.OrderBookPublisher
	pairs              map[string]*PairInfo
}

func NewMatchingService(
	engines map[string]*engine.Engine,
	tradeRepo interfaces.TradeRepository,
	orderClient interfaces.OrderClient,
	walletClient interfaces.WalletClient,
	tradePublisher pkgnats.TradePublisher,
	orderBookPublisher pkgnats.OrderBookPublisher,
	pairs map[string]*PairInfo,
) interfaces.MatchingService {
	return &matchingService{
		engines:            engines,
		tradeRepo:          tradeRepo,
		orderClient:        orderClient,
		walletClient:       walletClient,
		tradePublisher:     tradePublisher,
		orderBookPublisher: orderBookPublisher,
		pairs:              pairs,
	}
}

func (s *matchingService) SubmitOrder(ctx context.Context, order *engine.BookOrder) ([]*engine.Trade, error) {
	eng, ok := s.engines[order.PairID]
	if !ok {
		return nil, domain.ErrPairEngineNotFound
	}

	trades, err := eng.Submit(order)
	if err != nil {
		return nil, err
	}

	for _, trade := range trades {
		s.settleTrade(ctx, trade)
	}

	s.publishOrderBookSnapshot(order.PairID)

	return trades, nil
}

func (s *matchingService) GetOrderBook(_ context.Context, pairID string) (*engine.OrderBookSnapshot, error) {
	eng, ok := s.engines[pairID]
	if !ok {
		return nil, domain.ErrPairEngineNotFound
	}
	return eng.Snapshot()
}

func (s *matchingService) CancelOrder(ctx context.Context, pairID, orderID string) error {
	eng, ok := s.engines[pairID]
	if !ok {
		return domain.ErrPairEngineNotFound
	}

	if err := eng.Cancel(orderID); err != nil {
		return domain.ErrOrderNotInBook
	}

	s.publishOrderBookSnapshot(pairID)

	return nil
}

func (s *matchingService) publishOrderBookSnapshot(pairID string) {
	if s.orderBookPublisher == nil {
		return
	}
	eng, ok := s.engines[pairID]
	if !ok {
		return
	}
	snap, err := eng.Snapshot()
	if err != nil {
		slog.Warn("failed to get orderbook snapshot", "pair", pairID, "error", err.Error())
		return
	}

	bids := make([]pkgnats.OrderBookLevel, len(snap.Bids))
	for i, b := range snap.Bids {
		bids[i] = pkgnats.OrderBookLevel{Price: b.Price.String(), Quantity: b.Quantity.String(), OrderCount: b.OrderCount}
	}
	asks := make([]pkgnats.OrderBookLevel, len(snap.Asks))
	for i, a := range snap.Asks {
		asks[i] = pkgnats.OrderBookLevel{Price: a.Price.String(), Quantity: a.Quantity.String(), OrderCount: a.OrderCount}
	}

	if err := s.orderBookPublisher.PublishOrderBook(&pkgnats.OrderBookEvent{
		PairID: pairID,
		Bids:   bids,
		Asks:   asks,
	}); err != nil {
		slog.Warn("failed to publish orderbook snapshot", "pair", pairID, "error", err.Error())
	}
}

func (s *matchingService) settleTrade(ctx context.Context, trade *engine.Trade) {
	pair, ok := s.pairs[trade.PairID]
	if !ok {
		slog.ErrorContext(ctx, "pair not found for settlement", "pair_id", trade.PairID)
		return
	}

	quoteAmount := trade.Price.Mul(trade.Quantity)

	// 1. Settle buyer's locked quote asset (USDT locked -= quoteAmount)
	if err := s.walletClient.SettleTrade(ctx, trade.BuyerID, pair.QuoteAsset, quoteAmount); err != nil {
		slog.ErrorContext(ctx, "settlement failed: buyer settle quote",
			"trade_id", trade.ID, "buyer_id", trade.BuyerID, "error", err.Error())
	}

	// 2. Deposit base asset to buyer (BTC available += quantity)
	if err := s.walletClient.Deposit(ctx, trade.BuyerID, pair.BaseAsset, trade.Quantity); err != nil {
		slog.ErrorContext(ctx, "settlement failed: buyer deposit base",
			"trade_id", trade.ID, "buyer_id", trade.BuyerID, "error", err.Error())
	}

	// 3. Settle seller's locked base asset (BTC locked -= quantity)
	if err := s.walletClient.SettleTrade(ctx, trade.SellerID, pair.BaseAsset, trade.Quantity); err != nil {
		slog.ErrorContext(ctx, "settlement failed: seller settle base",
			"trade_id", trade.ID, "seller_id", trade.SellerID, "error", err.Error())
	}

	// 4. Deposit quote asset to seller (USDT available += quoteAmount)
	if err := s.walletClient.Deposit(ctx, trade.SellerID, pair.QuoteAsset, quoteAmount); err != nil {
		slog.ErrorContext(ctx, "settlement failed: seller deposit quote",
			"trade_id", trade.ID, "seller_id", trade.SellerID, "error", err.Error())
	}

	// 5. Update order fills
	s.updateOrderFills(ctx, trade)

	// 6. Persist trade
	s.persistTrade(ctx, trade)

	// 7. Publish NATS event
	s.publishTradeEvent(ctx, trade)
}

func (s *matchingService) updateOrderFills(ctx context.Context, trade *engine.Trade) {
	// Use trade quantity as the fill increment; order-svc SQL accumulates filled_qty
	// Determine status: if the engine order has no remaining, it's fully filled
	buyStatus := "partial"
	sellStatus := "partial"

	// Check if buy/sell orders are fully filled by inspecting the engine's BookOrder
	// The trade contains buyer/seller IDs but not remaining qty — use "partial" as default,
	// order-svc UpdateOrderFilledQty SQL guard ensures filled_qty <= quantity

	if err := s.orderClient.UpdateOrderFill(ctx, trade.BuyOrderID, trade.Quantity, buyStatus); err != nil {
		slog.ErrorContext(ctx, "failed to update buy order fill",
			"trade_id", trade.ID, "order_id", trade.BuyOrderID, "error", err.Error())
	}

	if err := s.orderClient.UpdateOrderFill(ctx, trade.SellOrderID, trade.Quantity, sellStatus); err != nil {
		slog.ErrorContext(ctx, "failed to update sell order fill",
			"trade_id", trade.ID, "order_id", trade.SellOrderID, "error", err.Error())
	}
}

func (s *matchingService) persistTrade(ctx context.Context, trade *engine.Trade) {
	domainTrade := &entities.Trade{
		ID:          trade.ID,
		PairID:      trade.PairID,
		BuyOrderID:  trade.BuyOrderID,
		SellOrderID: trade.SellOrderID,
		Price:       trade.Price,
		Quantity:    trade.Quantity,
		BuyerID:     trade.BuyerID,
		SellerID:    trade.SellerID,
		ExecutedAt:  trade.ExecutedAt,
	}

	if _, err := s.tradeRepo.Create(ctx, domainTrade); err != nil {
		slog.ErrorContext(ctx, "failed to persist trade",
			"trade_id", trade.ID, "error", err.Error())
	}
}

func (s *matchingService) publishTradeEvent(ctx context.Context, trade *engine.Trade) {
	if s.tradePublisher == nil {
		return
	}

	event := &pkgnats.TradeEvent{
		TradeID:     trade.ID,
		PairID:      trade.PairID,
		BuyOrderID:  trade.BuyOrderID,
		SellOrderID: trade.SellOrderID,
		Price:       trade.Price.String(),
		Quantity:    trade.Quantity.String(),
		BuyerID:     trade.BuyerID,
		SellerID:    trade.SellerID,
		ExecutedAt:  trade.ExecutedAt.Format(time.RFC3339),
	}

	if err := s.tradePublisher.PublishTrade(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to publish trade event",
			"trade_id", trade.ID, "error", err.Error())
	}
}

// DetermineOrderStatus calculates the order status based on filled vs total quantity.
func DetermineOrderStatus(filledQty, totalQty decimal.Decimal) string {
	if filledQty.GreaterThanOrEqual(totalQty) {
		return "filled"
	}
	return "partial"
}
