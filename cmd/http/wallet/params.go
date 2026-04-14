package wallet

// AssetIDParam documents the asset_id path parameter.
type AssetIDParam struct {
	AssetID string `params:"asset_id" required:"true"`
}
