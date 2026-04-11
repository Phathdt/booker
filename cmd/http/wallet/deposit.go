package wallet

import (
	"booker/modules/wallet/application/dto"
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Deposit godoc
// @Summary      Deposit funds to wallet
// @Tags         wallet
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.DepositDTO  true  "Deposit request"
// @Success      200   {object}  httpserver.Response{data=dto.WalletResponse}
// @Failure      400   {object}  httpserver.Response{error=object}
// @Failure      401   {object}  httpserver.Response{error=object}
// @Router       /api/v1/wallet/deposit [post]
func Deposit(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.DepositDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		userID := c.Locals("user_id").(string)
		w, err := walletSvc.Deposit(c.UserContext(), userID, req.AssetID, req.Amount)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toWalletResponse(w))
	}
}
