package users

import (
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetUser godoc
func GetUser(userSvc interfaces.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return fiber.NewError(fiber.StatusBadRequest, "User ID is required")
		}

		user, err := userSvc.GetByID(c.UserContext(), id)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toUserResponse(user))
	}
}
