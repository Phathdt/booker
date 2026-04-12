package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"booker/modules/notification/domain/entities"
	"booker/modules/notification/domain/interfaces"
	pkgnats "booker/pkg/nats"

	"github.com/nats-io/nats.go"
)

// EventHandler maps NATS events to notifications.
type EventHandler struct {
	notifSvc interfaces.NotificationService
}

func NewEventHandler(notifSvc interfaces.NotificationService) *EventHandler {
	return &EventHandler{notifSvc: notifSvc}
}

// Handle routes a NATS message to the appropriate event handler.
func (h *EventHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	subject := msg.Subject
	switch {
	case strings.HasPrefix(subject, "trades."):
		return h.handleTradeEvent(ctx, msg.Data)
	case strings.HasPrefix(subject, "orders."):
		return h.handleOrderEvent(ctx, msg.Data)
	case strings.HasPrefix(subject, "wallets."):
		return h.handleWalletEvent(ctx, msg.Data)
	default:
		return nil
	}
}

func (h *EventHandler) handleTradeEvent(ctx context.Context, data []byte) error {
	var event pkgnats.TradeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal trade event: %w", err)
	}

	// Notify both buyer and seller independently — one failure should not block the other
	users := []struct {
		id   string
		side string
	}{
		{event.BuyerID, "bought"},
		{event.SellerID, "sold"},
	}
	var errs []error
	for _, u := range users {
		notif := &entities.Notification{
			UserID:   u.id,
			EventKey: fmt.Sprintf("trade_%s_%s", event.TradeID, u.id),
			Type:     entities.TypeTradeExecuted,
			Title:    "Trade Executed",
			Body:     fmt.Sprintf("You %s %s %s at %s", u.side, event.Quantity, event.PairID, event.Price),
			Metadata: map[string]string{
				"trade_id": event.TradeID,
				"pair_id":  event.PairID,
			},
		}
		if _, err := h.notifSvc.CreateNotification(ctx, notif); err != nil {
			errs = append(errs, fmt.Errorf("user %s: %w", u.id, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("trade notification errors: %v", errs)
	}
	return nil
}

func (h *EventHandler) handleOrderEvent(ctx context.Context, data []byte) error {
	var event pkgnats.OrderEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal order event: %w", err)
	}

	var title, body string
	switch event.Status {
	case "filled":
		title = "Order Filled"
		body = fmt.Sprintf("Your %s order for %s %s has been fully filled", event.Side, event.Quantity, event.PairID)
	case "cancelled":
		title = "Order Cancelled"
		body = fmt.Sprintf("Your %s order for %s %s has been cancelled", event.Side, event.Quantity, event.PairID)
	default:
		return nil
	}

	notifType := entities.TypeOrderFilled
	if event.Status == "cancelled" {
		notifType = entities.TypeOrderCancelled
	}

	notif := &entities.Notification{
		UserID:   event.UserID,
		EventKey: fmt.Sprintf("order_%s_%s", event.OrderID, event.Status),
		Type:     notifType,
		Title:    title,
		Body:     body,
		Metadata: map[string]string{
			"order_id": event.OrderID,
			"pair_id":  event.PairID,
			"side":     event.Side,
		},
	}
	_, createErr := h.notifSvc.CreateNotification(ctx, notif)
	return createErr
}

func (h *EventHandler) handleWalletEvent(ctx context.Context, data []byte) error {
	var event pkgnats.WalletEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal wallet event: %w", err)
	}

	var title, body string
	var notifType entities.NotificationType
	switch event.Action {
	case "deposit":
		title = "Deposit Confirmed"
		body = fmt.Sprintf("Your deposit of %s %s has been confirmed", event.Amount, event.Asset)
		notifType = entities.TypeDepositConfirmed
	case "withdrawal":
		title = "Withdrawal Confirmed"
		body = fmt.Sprintf("Your withdrawal of %s %s has been confirmed", event.Amount, event.Asset)
		notifType = entities.TypeWithdrawalConfirmed
	default:
		return nil
	}

	notif := &entities.Notification{
		UserID:   event.UserID,
		EventKey: fmt.Sprintf("wallet_%s", event.TxID),
		Type:     notifType,
		Title:    title,
		Body:     body,
		Metadata: map[string]string{
			"asset":  event.Asset,
			"amount": event.Amount,
			"tx_id":  event.TxID,
		},
	}
	_, createErr := h.notifSvc.CreateNotification(ctx, notif)
	return createErr
}
