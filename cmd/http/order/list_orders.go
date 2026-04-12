package order

import (
	"booker/modules/order/application/dto"
	"booker/modules/order/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// ListOrders godoc
// @Summary      List orders for current user
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        pair_id  query  string  false  "Filter by trading pair (e.g. BTC_USDT)"
// @Param        status   query  string  false  "Filter by status (new, partial, filled, cancelled)"
// @Param        limit    query  int     false  "Page size (default 20, max 100)"
// @Param        offset   query  int     false  "Offset (default 0, max 10000)"
// @Success      200  {object}  httpserver.Response{data=dto.OrderListResponse}
// @Failure      400  {object}  httpserver.Response{error=object}
// @Failure      401  {object}  httpserver.Response{error=object}
// @Router       /api/v1/orders [get]
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
