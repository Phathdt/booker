package market

// PairPathParam documents the pair path parameter.
type PairPathParam struct {
	Pair string `params:"pair" required:"true"`
}

// TradesQueryParam documents the trades endpoint query parameters.
type TradesQueryParam struct {
	Pair  string `params:"pair"  required:"true"`
	Limit int    `query:"limit"`
}

// OrderBookQueryParam documents the orderbook endpoint query parameters.
type OrderBookQueryParam struct {
	Pair  string `params:"pair"  required:"true"`
	Depth int    `query:"depth"`
}
