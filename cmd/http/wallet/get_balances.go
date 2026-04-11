package wallet

import (
	"booker/modules/wallet/application/dto"
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetBalances godoc
// @Summary      Get all wallet balances for current user
// @Tags         wallet
// @Security     BearerAuth
// @Success      200  {object}  httpserver.Response{data=dto.WalletListResponse}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Router       /api/v1/wallet [get]
func GetBalances(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		wallets, err := walletSvc.GetBalances(c.UserContext(), userID)
		if err != nil {
			return err
		}

		items := make([]dto.WalletResponse, len(wallets))
		for i, w := range wallets {
			items[i] = toWalletResponse(w)
		}

		return httpserver.OK(c, dto.WalletListResponse{Wallets: items})
	}
}
