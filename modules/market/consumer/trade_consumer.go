package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"booker/modules/market/ticker"
	"booker/modules/market/trades"
	"booker/modules/market/ws"
	pkgnats "booker/pkg/nats"

	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
)

// TradeConsumer subscribes to NATS trade events and fans out to ticker, trades, and WS hub.
type TradeConsumer struct {
	tickers map[string]*ticker.Aggregator
	trades  map[string]*trades.RecentTrades
	hub     *ws.Hub
	sub     *nats.Subscription
}

func NewTradeConsumer(
	tickers map[string]*ticker.Aggregator,
	recentTrades map[string]*trades.RecentTrades,
	hub *ws.Hub,
) *TradeConsumer {
	return &TradeConsumer{
		tickers: tickers,
		trades:  recentTrades,
		hub:     hub,
	}
}

// Start begins consuming trade events from NATS JetStream.
func (c *TradeConsumer) Start(ctx context.Context, js nats.JetStreamContext) error {
	sub, err := js.Subscribe("trades.>", func(msg *nats.Msg) {
		var event pkgnats.TradeEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			slog.Error("failed to unmarshal trade event", "error", err.Error())
			msg.Ack()
			return
		}

		c.processTrade(&event)
		msg.Ack()
	}, nats.Durable("market-svc"), nats.DeliverNew(), nats.AckExplicit())
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

func (c *TradeConsumer) processTrade(event *pkgnats.TradeEvent) {
	price, err := decimal.NewFromString(event.Price)
	if err != nil {
		slog.Error("invalid trade price", "trade_id", event.TradeID, "price", event.Price)
		return
	}
	qty, err := decimal.NewFromString(event.Quantity)
	if err != nil {
		slog.Error("invalid trade quantity", "trade_id", event.TradeID, "quantity", event.Quantity)
		return
	}

	ts, err := time.Parse(time.RFC3339, event.ExecutedAt)
	if err != nil {
		slog.Error("invalid trade timestamp", "trade_id", event.TradeID, "executed_at", event.ExecutedAt)
		ts = time.Now()
	}
	tsMs := ts.UnixMilli()

	// Update ticker
	if agg, ok := c.tickers[event.PairID]; ok {
		agg.AddTrade(price, qty, ts)
	}

	// Add to recent trades
	if rt, ok := c.trades[event.PairID]; ok {
		rt.Add(trades.TradeInfo{
			TradeID:   event.TradeID,
			Price:     event.Price,
			Quantity:  event.Quantity,
			BuyerID:   event.BuyerID,
			SellerID:  event.SellerID,
			Timestamp: tsMs,
		})
	}

	// Broadcast to WS hub
	if c.hub != nil {
		// Broadcast trade event
		c.hub.BroadcastTrade(event.PairID, ws.TradeData{
			TradeID:   event.TradeID,
			Price:     event.Price,
			Quantity:  event.Quantity,
			Timestamp: tsMs,
		})

		// Broadcast updated ticker
		if agg, ok := c.tickers[event.PairID]; ok {
			t := agg.GetTicker()
			c.hub.BroadcastTicker(event.PairID, ws.TickerData{
				Open:      t.Open.String(),
				High:      t.High.String(),
				Low:       t.Low.String(),
				Close:     t.Close.String(),
				Volume:    t.Volume.String(),
				ChangePct: t.ChangePct.String(),
				LastPrice: t.LastPrice.String(),
				Timestamp: t.Timestamp,
			})
		}
	}
}
