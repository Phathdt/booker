package order

import (
	"booker/modules/order/application/dto"
	"booker/modules/order/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// ListOrders godoc
func ListOrders(orderSvc interfaces.OrderService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.ListOrdersDTO
		if err := c.QueryParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters")
		}
		if err := httpserver.ValidateStruct(c, &req); err != nil {
			return err
		}

		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		orders, err := orderSvc.ListOrders(c.UserContext(), userID, &req)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toOrderListResponse(orders))
	}
}
