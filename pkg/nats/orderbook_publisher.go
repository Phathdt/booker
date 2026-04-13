package nats

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// OrderBookLevel represents a single price level in the order book snapshot.
type OrderBookLevel struct {
	Price      string `json:"price"`
	Quantity   string `json:"quantity"`
	OrderCount int    `json:"order_count"`
}

// OrderBookEvent is the NATS message for order book updates.
type OrderBookEvent struct {
	PairID string           `json:"pair_id"`
	Bids   []OrderBookLevel `json:"bids"`
	Asks   []OrderBookLevel `json:"asks"`
}

// OrderBookPublisher publishes order book snapshots to NATS.
type OrderBookPublisher interface {
	PublishOrderBook(event *OrderBookEvent) error
}

type orderBookPublisher struct {
	js nats.JetStreamContext
}

// NewOrderBookPublisher creates an OrderBookPublisher backed by NATS JetStream.
func NewOrderBookPublisher(js nats.JetStreamContext) OrderBookPublisher {
	return &orderBookPublisher{js: js}
}

func (p *orderBookPublisher) PublishOrderBook(event *OrderBookEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal orderbook event: %w", err)
	}

	subject := fmt.Sprintf("orderbook.%s.updated", event.PairID)
	if _, err := p.js.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish orderbook event: %w", err)
	}
	return nil
}
