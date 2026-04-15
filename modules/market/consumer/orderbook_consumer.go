package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"booker/modules/market/ws"
	pkgnats "booker/pkg/nats"

	"github.com/nats-io/nats.go"
)

// OrderBookConsumer subscribes to NATS orderbook events and broadcasts to WS hub.
type OrderBookConsumer struct {
	hub *ws.Hub
	sub *nats.Subscription
}

func NewOrderBookConsumer(hub *ws.Hub) *OrderBookConsumer {
	return &OrderBookConsumer{hub: hub}
}

// Start begins consuming orderbook events from NATS JetStream.
func (c *OrderBookConsumer) Start(ctx context.Context, js nats.JetStreamContext) error {
	// Ephemeral consumer — every market-svc instance gets all messages (broadcast, not load-balanced)
	sub, err := js.Subscribe("orderbook.>", func(msg *nats.Msg) {
		var event pkgnats.OrderBookEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			slog.Error("failed to unmarshal orderbook event", "error", err.Error())
			msg.Ack()
			return
		}

		bids := make([]ws.OrderBookLevelData, len(event.Bids))
		for i, b := range event.Bids {
			bids[i] = ws.OrderBookLevelData{Price: b.Price, Quantity: b.Quantity, OrderCount: b.OrderCount}
		}
		asks := make([]ws.OrderBookLevelData, len(event.Asks))
		for i, a := range event.Asks {
			asks[i] = ws.OrderBookLevelData{Price: a.Price, Quantity: a.Quantity, OrderCount: a.OrderCount}
		}

		c.hub.BroadcastOrderBook(event.PairID, ws.OrderBookData{
			Bids: bids,
			Asks: asks,
		})

		msg.Ack()
	}, nats.DeliverLast(), nats.AckExplicit())
	if err != nil {
		return err
	}
	c.sub = sub

	go func() {
		<-ctx.Done()
		c.sub.Unsubscribe()
	}()

	return nil
}
