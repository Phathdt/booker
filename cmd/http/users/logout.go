package users

import (
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Logout godoc
// @Summary      Logout (revoke all tokens)
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  httpserver.Response{data=dto.MessageResponse}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Router       /api/v1/auth/logout [post]
func Logout(uc *usecases.LogoutUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		if err := uc.Execute(c.UserContext(), userID); err != nil {
			return err
		}
		return httpserver.OK(c, dto.MessageResponse{Message: "logged out"})
	}
}
