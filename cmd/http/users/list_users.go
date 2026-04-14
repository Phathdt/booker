package users

import (
	"strconv"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// ListUsers godoc
func ListUsers(userSvc interfaces.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := strconv.Atoi(c.Query("limit", "20"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid limit parameter")
		}
		offset, err := strconv.Atoi(c.Query("offset", "0"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid offset parameter")
		}

		if limit <= 0 || limit > 100 {
			limit = 20
		}
		if offset < 0 {
			offset = 0
		}

		users, total, err := userSvc.List(c.UserContext(), limit, offset)
		if err != nil {
			return err
		}

		items := make([]dto.UserResponse, len(users))
		for i, u := range users {
			items[i] = toUserResponse(u)
		}

		return httpserver.OK(c, dto.UserListResponse{Users: items, Total: total})
	}
}
