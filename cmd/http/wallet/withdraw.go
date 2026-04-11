package wallet

import (
	"booker/modules/wallet/application/dto"
	"booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Withdraw godoc
// @Summary      Withdraw funds from wallet
// @Tags         wallet
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.WithdrawDTO  true  "Withdraw request"
// @Success      200   {object}  httpserver.Response{data=dto.WalletResponse}
// @Failure      400   {object}  httpserver.Response{error=object}
// @Failure      401   {object}  httpserver.Response{error=object}
// @Router       /api/v1/wallet/withdraw [post]
func Withdraw(walletSvc interfaces.WalletService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.WithdrawDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		userID := c.Locals("user_id").(string)
		w, err := walletSvc.Withdraw(c.UserContext(), userID, req.AssetID, req.Amount)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toWalletResponse(w))
	}
}
