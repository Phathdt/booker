package entities

import "time"

type NotificationType string

const (
	TypeTradeExecuted       NotificationType = "trade_executed"
	TypeOrderFilled         NotificationType = "order_filled"
	TypeOrderCancelled      NotificationType = "order_cancelled"
	TypeDepositConfirmed    NotificationType = "deposit_confirmed"
	TypeWithdrawalConfirmed NotificationType = "withdrawal_confirmed"
)

type Notification struct {
	ID        string
	UserID    string
	EventKey  string // dedup: "trade_{id}_{user}", "order_{id}_{status}", "wallet_{txid}"
	Type      NotificationType
	Title     string
	Body      string
	Metadata  map[string]string
	IsRead    bool
	CreatedAt time.Time
}
