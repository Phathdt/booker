package order

import (
	"booker/modules/order/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CancelOrder godoc
func CancelOrder(orderSvc interfaces.OrderService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orderID := c.Params("id")
		if _, err := uuid.Parse(orderID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid order ID format")
		}

		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		order, err := orderSvc.CancelOrder(c.UserContext(), userID, orderID)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toOrderResponse(order))
	}
}
