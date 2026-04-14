package market

// PairInfo holds trading pair metadata for REST responses.
type PairInfo struct {
	ID         string
	BaseAsset  string
	QuoteAsset string
	MinQty     string
	TickSize   string
}

type PairResponse struct {
	ID         string `json:"id"          required:"true" example:"BTC_USDT"`
	BaseAsset  string `json:"base_asset"  required:"true" example:"BTC"`
	QuoteAsset string `json:"quote_asset" required:"true" example:"USDT"`
	MinQty     string `json:"min_qty"     required:"true" example:"0.00001"`
	TickSize   string `json:"tick_size"   required:"true" example:"0.01"`
}

type TickerResponse struct {
	Pair      string `json:"pair"       required:"true" example:"BTC_USDT"`
	Open      string `json:"open"       required:"true" example:"50000.00"`
	High      string `json:"high"       required:"true" example:"51000.00"`
	Low       string `json:"low"        required:"true" example:"49000.00"`
	Close     string `json:"close"      required:"true" example:"50500.00"`
	Volume    string `json:"volume"     required:"true" example:"1234.56"`
	ChangePct string `json:"change_pct" required:"true" example:"1.00"`
	LastPrice string `json:"last_price" required:"true" example:"50500.00"`
	Timestamp int64  `json:"timestamp"  required:"true" example:"1744444444000"`
}

type TradeResponse struct {
	TradeID   string `json:"trade_id"  required:"true" example:"550e8400-e29b-41d4-a716-446655440000"`
	Price     string `json:"price"     required:"true" example:"50000.00"`
	Quantity  string `json:"quantity"  required:"true" example:"0.5"`
	Timestamp int64  `json:"timestamp" required:"true" example:"1744444444000"`
}
