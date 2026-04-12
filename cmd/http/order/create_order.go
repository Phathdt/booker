package order

import (
	"booker/modules/order/application/dto"
	"booker/modules/order/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// CreateOrder godoc
// @Summary      Create a new limit order
// @Tags         orders
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateOrderDTO  true  "Create order request"
// @Success      201   {object}  httpserver.Response{data=dto.OrderResponse}
// @Failure      400   {object}  httpserver.Response{error=object}
// @Failure      401   {object}  httpserver.Response{error=object}
// @Router       /api/v1/orders [post]
func CreateOrder(orderSvc interfaces.OrderService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.CreateOrderDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		userID := c.Locals("user_id").(string)
		order, err := orderSvc.CreateOrder(c.UserContext(), userID, &req)
		if err != nil {
			return err
		}

		return httpserver.Created(c, toOrderResponse(order))
	}
}
