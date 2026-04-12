package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// OrderEvent represents an order status change event.
type OrderEvent struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	PairID    string `json:"pair_id"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	FilledQty string `json:"filled_qty"`
	Status    string `json:"status"` // "filled", "cancelled"
	UpdatedAt string `json:"updated_at"`
}

// OrderPublisher publishes order events to NATS JetStream.
type OrderPublisher interface {
	PublishOrderUpdate(ctx context.Context, event *OrderEvent) error
}

type orderPublisher struct {
	js nats.JetStreamContext
}

func NewOrderPublisher(js nats.JetStreamContext) OrderPublisher {
	return &orderPublisher{js: js}
}

func (p *orderPublisher) PublishOrderUpdate(ctx context.Context, event *OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order event: %w", err)
	}

	subject := fmt.Sprintf("orders.%s.%s", event.UserID, event.Status)
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s_%s", event.OrderID, event.Status))

	if _, err := p.js.PublishMsg(msg); err != nil {
		return fmt.Errorf("failed to publish order event: %w", err)
	}
	return nil
}
