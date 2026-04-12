package consumer

import (
	"context"
	"time"

	"booker/pkg/logger"

	"github.com/nats-io/nats.go"
)

// NATSConsumer subscribes to JetStream streams and processes events.
type NATSConsumer struct {
	js      nats.JetStreamContext
	handler *EventHandler
	log     logger.Logger
	subs    []*nats.Subscription
	cancel  context.CancelFunc
}

func NewNATSConsumer(js nats.JetStreamContext, handler *EventHandler, log logger.Logger) *NATSConsumer {
	return &NATSConsumer{
		js:      js,
		handler: handler,
		log:     log,
	}
}

// Start creates durable consumers and begins processing events.
func (c *NATSConsumer) Start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)

	streams := []struct {
		subject  string
		consumer string
	}{
		{"trades.>", "notif-trades-consumer"},
		{"orders.>", "notif-orders-consumer"},
		{"wallets.>", "notif-wallets-consumer"},
	}

	for _, s := range streams {
		sub, err := c.js.PullSubscribe(s.subject, s.consumer,
			nats.AckExplicit(),
			nats.MaxDeliver(5),
		)
		if err != nil {
			c.log.With("subject", s.subject, "error", err.Error()).Error("failed to subscribe")
			continue
		}
		c.subs = append(c.subs, sub)
		go c.processMessages(ctx, sub, s.subject)
		c.log.With("subject", s.subject, "consumer", s.consumer).Info("NATS consumer started")
	}
}

func (c *NATSConsumer) processMessages(ctx context.Context, sub *nats.Subscription, stream string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
			if err != nil {
				if err == nats.ErrTimeout || err == context.DeadlineExceeded {
					continue
				}
				c.log.With("stream", stream, "error", err.Error()).Error("failed to fetch messages")
				// Backoff on persistent errors to avoid tight loop
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				continue
			}
			for _, msg := range msgs {
				if err := c.handler.Handle(ctx, msg); err != nil {
					c.log.With("stream", stream, "error", err.Error()).Error("failed to handle event")
					// NakWithDelay to avoid tight retry loop on transient failures
					if nakErr := msg.NakWithDelay(5 * time.Second); nakErr != nil {
						c.log.With("stream", stream, "error", nakErr.Error()).Warn("failed to nak message")
					}
				} else {
					if ackErr := msg.Ack(); ackErr != nil {
						c.log.With("stream", stream, "error", ackErr.Error()).Warn("failed to ack message")
					}
				}
			}
		}
	}
}

// Stop drains all subscriptions and cancels processing.
func (c *NATSConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	for _, sub := range c.subs {
		if err := sub.Drain(); err != nil {
			c.log.With("error", err.Error()).Warn("failed to drain NATS subscription")
		}
	}
}
