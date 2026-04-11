package wallet

import (
	"time"

	"booker/modules/wallet/application/dto"
	"booker/modules/wallet/domain/entities"
)

func toWalletResponse(w *entities.Wallet) dto.WalletResponse {
	return dto.WalletResponse{
		ID:        w.ID,
		UserID:    w.UserID,
		AssetID:   w.AssetID,
		Available: w.Available.String(),
		Locked:    w.Locked.String(),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}
