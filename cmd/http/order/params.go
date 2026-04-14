package order

// OrderIDParam documents the order ID path parameter.
type OrderIDParam struct {
	ID string `params:"id" required:"true"`
}

// ListOrdersParam documents the list orders query parameters.
type ListOrdersParam struct {
	PairID string `query:"pair_id"`
	Status string `query:"status"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}
