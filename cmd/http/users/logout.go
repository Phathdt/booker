package users

import (
	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Logout godoc
func Logout(cfg *config.Config, uc *usecases.LogoutUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		if err := uc.Execute(c.UserContext(), userID); err != nil {
			return err
		}
		clearRefreshTokenCookie(c, cfg)
		return httpserver.OK(c, dto.MessageResponse{Message: "logged out"})
	}
}
