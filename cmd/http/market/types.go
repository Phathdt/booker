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
	ID         string `json:"id"          example:"BTC_USDT"`
	BaseAsset  string `json:"base_asset"  example:"BTC"`
	QuoteAsset string `json:"quote_asset" example:"USDT"`
	MinQty     string `json:"min_qty"     example:"0.00001"`
	TickSize   string `json:"tick_size"   example:"0.01"`
}

type TickerResponse struct {
	Pair      string `json:"pair"       example:"BTC_USDT"`
	Open      string `json:"open"       example:"50000.00"`
	High      string `json:"high"       example:"51000.00"`
	Low       string `json:"low"        example:"49000.00"`
	Close     string `json:"close"      example:"50500.00"`
	Volume    string `json:"volume"     example:"1234.56"`
	ChangePct string `json:"change_pct" example:"1.00"`
	LastPrice string `json:"last_price" example:"50500.00"`
	Timestamp int64  `json:"timestamp"  example:"1744444444000"`
}

type TradeResponse struct {
	TradeID   string `json:"trade_id"  example:"550e8400-e29b-41d4-a716-446655440000"`
	Price     string `json:"price"     example:"50000.00"`
	Quantity  string `json:"quantity"  example:"0.5"`
	Timestamp int64  `json:"timestamp" example:"1744444444000"`
}
