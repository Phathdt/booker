package ws

// SubscribeMsg is the client → server message for subscribe/unsubscribe.
type SubscribeMsg struct {
	Op      string `json:"op"`      // "subscribe" or "unsubscribe"
	Channel string `json:"channel"` // "ticker" or "trades"
	Pair    string `json:"pair"`    // "BTC_USDT"
}

// WSMessage is the server → client message.
type WSMessage struct {
	Type string      `json:"type"`
	Pair string      `json:"pair,omitempty"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"msg,omitempty"`
}

// TickerData is the payload for ticker updates.
type TickerData struct {
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
	ChangePct string `json:"change_pct"`
	LastPrice string `json:"last_price"`
	Timestamp int64  `json:"ts"`
}

// TradeData is the payload for trade events.
type TradeData struct {
	TradeID   string `json:"trade_id"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	Timestamp int64  `json:"ts"`
}

// OrderBookLevelData is a single price level in the order book.
type OrderBookLevelData struct {
	Price      string `json:"price"`
	Quantity   string `json:"quantity"`
	OrderCount int    `json:"order_count"`
}

// OrderBookData is the payload for order book updates.
type OrderBookData struct {
	Bids []OrderBookLevelData `json:"bids"`
	Asks []OrderBookLevelData `json:"asks"`
}
