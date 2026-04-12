package users

import (
	_ "booker/modules/users/application/dto" // swagger
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetMe godoc
// @Summary      Get current authenticated user
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  httpserver.Response{data=dto.UserResponse}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Router       /api/v1/auth/me [get]
func GetMe(userSvc interfaces.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		user, err := userSvc.GetByID(c.UserContext(), userID)
		if err != nil {
			return err
		}
		return httpserver.OK(c, toUserResponse(user))
	}
}
