package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// WalletEvent represents a wallet action event.
type WalletEvent struct {
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Amount    string `json:"amount"`
	Action    string `json:"action"` // "deposit", "withdrawal"
	TxID      string `json:"tx_id"`
	CreatedAt string `json:"created_at"`
}

// WalletPublisher publishes wallet events to NATS JetStream.
type WalletPublisher interface {
	PublishWalletAction(ctx context.Context, event *WalletEvent) error
}

type walletPublisher struct {
	js nats.JetStreamContext
}

func NewWalletPublisher(js nats.JetStreamContext) WalletPublisher {
	return &walletPublisher{js: js}
}

func (p *walletPublisher) PublishWalletAction(ctx context.Context, event *WalletEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet event: %w", err)
	}

	subject := fmt.Sprintf("wallets.%s.%s", event.UserID, event.Action)
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Nats-Msg-Id", event.TxID)

	if _, err := p.js.PublishMsg(msg); err != nil {
		return fmt.Errorf("failed to publish wallet event: %w", err)
	}
	return nil
}
