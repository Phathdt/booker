package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// EnsureStreams creates all JetStream streams if they don't exist.
func EnsureStreams(js nats.JetStreamContext) error {
	streams := []nats.StreamConfig{
		{Name: "TRADES", Subjects: []string{"trades.>"}, Storage: nats.FileStorage},
		{Name: "ORDERS", Subjects: []string{"orders.>"}, Storage: nats.FileStorage},
		{Name: "WALLETS", Subjects: []string{"wallets.>"}, Storage: nats.FileStorage},
		{Name: "ORDERBOOK", Subjects: []string{"orderbook.>"}, Storage: nats.MemoryStorage, MaxMsgsPerSubject: 1, Discard: nats.DiscardOld},
	}
	for _, cfg := range streams {
		if _, err := js.AddStream(&cfg); err != nil {
			return fmt.Errorf("failed to create %s stream: %w", cfg.Name, err)
		}
	}
	return nil
}
