package wallet

import (
	_ "booker/modules/wallet/application/dto" // swagger
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetBalance godoc
// @Summary      Get wallet balance for a specific asset
// @Tags         wallet
// @Security     BearerAuth
// @Param        asset_id  path  string  true  "Asset ID (e.g. BTC, USDT)"
// @Success      200  {object}  httpserver.Response{data=dto.WalletResponse}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Router       /api/v1/wallet/{asset_id} [get]
func GetBalance(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
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
