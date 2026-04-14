package dto

import "github.com/shopspring/decimal"

type CreateOrderDTO struct {
	PairID   string          `json:"pair_id"  validate:"required" required:"true"`
	Side     string          `json:"side"     validate:"required,oneof=buy sell" required:"true"`
	Type     string          `json:"type"     validate:"required,oneof=limit" required:"true"`
	Price    decimal.Decimal `json:"price"    validate:"required" required:"true"`
	Quantity decimal.Decimal `json:"quantity" validate:"required" required:"true"`
}

type ListOrdersDTO struct {
	PairID string `query:"pair_id"`
	Status string `query:"status"  validate:"omitempty,oneof=new partial filled cancelled"`
	Limit  int32  `query:"limit"   validate:"omitempty,min=1,max=100"`
	Offset int32  `query:"offset"  validate:"omitempty,min=0,max=10000"`
}

type OrderResponse struct {
	ID        string `json:"id"         required:"true" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string `json:"user_id"    required:"true" example:"550e8400-e29b-41d4-a716-446655440001"`
	PairID    string `json:"pair_id"    required:"true" example:"BTC_USDT"`
	Side      string `json:"side"       required:"true" example:"buy"`
	Type      string `json:"type"       required:"true" example:"limit"`
	Price     string `json:"price"      required:"true" example:"50000.00"`
	Quantity  string `json:"quantity"   required:"true" example:"0.5"`
	FilledQty string `json:"filled_qty" required:"true" example:"0"`
	Status    string `json:"status"     required:"true" example:"new"`
	CreatedAt string `json:"created_at" required:"true" example:"2026-04-12T00:00:00Z"`
	UpdatedAt string `json:"updated_at" required:"true" example:"2026-04-12T00:00:00Z"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders" required:"true"`
}
