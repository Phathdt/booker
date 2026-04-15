package wallet

import (
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetBalance godoc
func GetBalance(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		assetID := c.Params("asset_id")
		if assetID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Asset ID is required")
		}

		w, err := walletSvc.GetBalance(c.UserContext(), userID, assetID)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toWalletResponse(w))
	}
}
