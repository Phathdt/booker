package dto

import "github.com/shopspring/decimal"

// DepositDTO holds deposit request input.
type DepositDTO struct {
	AssetID string          `json:"asset_id" validate:"required" required:"true"`
	Amount  decimal.Decimal `json:"amount"   validate:"required" required:"true"`
}

// WithdrawDTO holds withdraw request input.
type WithdrawDTO struct {
	AssetID string          `json:"asset_id" validate:"required" required:"true"`
	Amount  decimal.Decimal `json:"amount"   validate:"required" required:"true"`
}

// WalletResponse represents a wallet in API responses.
type WalletResponse struct {
	ID        string `json:"id"         required:"true" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string `json:"user_id"    required:"true" example:"550e8400-e29b-41d4-a716-446655440001"`
	AssetID   string `json:"asset_id"   required:"true" example:"USDT"`
	Available string `json:"available"  required:"true" example:"1000.00"`
	Locked    string `json:"locked"     required:"true" example:"0.00"`
	UpdatedAt string `json:"updated_at" required:"true" example:"2026-04-12T00:00:00Z"`
}

// WalletListResponse represents a list of wallets.
type WalletListResponse struct {
	Wallets []WalletResponse `json:"wallets" required:"true"`
}
