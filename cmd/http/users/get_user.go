package users

import (
	_ "booker/modules/users/application/dto" // swagger
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetUser godoc
// @Summary      Get user by ID
// @Tags         users
// @Security     BearerAuth
// @Param        id  path  string  true  "User ID"
// @Success      200  {object}  httpserver.Response{data=dto.UserResponse}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Failure      404  {object}  httpserver.Response{error=object}
// @Router       /api/v1/users/{id} [get]
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
