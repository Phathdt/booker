package order

import (
	orderInterfaces "booker/modules/order/domain/interfaces"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes sets up order HTTP routes on the Fiber app.
func RegisterRoutes(app *fiber.App, orderSvc orderInterfaces.OrderService, tokenSvc interfaces.TokenService) {
	o := app.Group("/api/v1/orders", httpserver.AuthMiddleware(tokenSvc))

	o.Post("/", CreateOrder(orderSvc))
	o.Get("/", ListOrders(orderSvc))
	o.Get("/:id", GetOrder(orderSvc))
	o.Delete("/:id", CancelOrder(orderSvc))
}
