package users

import (
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetMe godoc
func GetMe(userSvc interfaces.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		user, err := userSvc.GetByID(c.UserContext(), userID)
		if err != nil {
			return err
		}
		return httpserver.OK(c, toUserResponse(user))
	}
}
