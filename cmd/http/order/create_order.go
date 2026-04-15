package order

import (
	"booker/modules/order/application/dto"
	"booker/modules/order/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// CreateOrder godoc
func CreateOrder(orderSvc interfaces.OrderService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.CreateOrderDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		order, err := orderSvc.CreateOrder(c.UserContext(), userID, &req)
		if err != nil {
			return err
		}

		return httpserver.Created(c, toOrderResponse(order))
	}
}
