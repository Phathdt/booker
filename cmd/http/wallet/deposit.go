package wallet

import (
	"booker/modules/wallet/application/dto"
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Deposit godoc
func Deposit(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.DepositDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		w, err := walletSvc.Deposit(c.UserContext(), userID, req.AssetID, req.Amount)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toWalletResponse(w))
	}
}
