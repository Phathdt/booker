package users

import (
	"strconv"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// ListUsers godoc
// @Summary      List users
// @Tags         users
// @Security     BearerAuth
// @Param        limit   query  int  false  "Limit"   default(20)
// @Param        offset  query  int  false  "Offset"  default(0)
// @Success      200  {object}  httpserver.Response{data=dto.UserListResponse}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Router       /api/v1/users [get]
func ListUsers(userSvc interfaces.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, _ := strconv.Atoi(c.Query("limit", "20"))
		offset, _ := strconv.Atoi(c.Query("offset", "0"))

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
