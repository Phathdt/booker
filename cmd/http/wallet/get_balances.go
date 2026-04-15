package wallet

import (
	"booker/modules/wallet/application/dto"
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetBalances godoc
func GetBalances(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}

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
