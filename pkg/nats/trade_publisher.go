package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// TradeEvent represents a trade event published to NATS.
type TradeEvent struct {
	TradeID     string `json:"trade_id"`
	PairID      string `json:"pair_id"`
	BuyOrderID  string `json:"buy_order_id"`
	SellOrderID string `json:"sell_order_id"`
	Price       string `json:"price"`
	Quantity    string `json:"quantity"`
	BuyerID     string `json:"buyer_id"`
	SellerID    string `json:"seller_id"`
	ExecutedAt  string `json:"executed_at"`
}

// TradePublisher publishes trade events to NATS JetStream.
type TradePublisher interface {
	PublishTrade(ctx context.Context, event *TradeEvent) error
}

type jetStreamPublisher struct {
	js nats.JetStreamContext
}

// NewTradePublisher creates a TradePublisher backed by NATS JetStream.
func NewTradePublisher(js nats.JetStreamContext) TradePublisher {
	return &jetStreamPublisher{js: js}
}

// EnsureStream creates the TRADES stream if it doesn't exist.
func EnsureStream(js nats.JetStreamContext) error {
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     "TRADES",
		Subjects: []string{"trades.>"},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to create TRADES stream: %w", err)
	}
	return nil
}

func (p *jetStreamPublisher) PublishTrade(ctx context.Context, event *TradeEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal trade event: %w", err)
	}

	subject := fmt.Sprintf("trades.%s.executed", event.PairID)
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	// Dedup via Nats-Msg-Id
	msg.Header.Set("Nats-Msg-Id", event.TradeID)

	if _, err := p.js.PublishMsg(msg); err != nil {
		return fmt.Errorf("failed to publish trade event: %w", err)
	}
	return nil
}
