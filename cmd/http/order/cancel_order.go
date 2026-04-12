package order

import (
	_ "booker/modules/order/application/dto" // swagger
	"booker/modules/order/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CancelOrder godoc
// @Summary      Cancel an order
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Order ID (UUID)"
// @Success      200  {object}  httpserver.Response{data=dto.OrderResponse}
// @Failure      400  {object}  httpserver.Response{error=object}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Failure      404  {object}  httpserver.Response{error=object}
// @Router       /api/v1/orders/{id} [delete]
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
