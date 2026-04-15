package order

import (
	orderDTO "booker/modules/order/application/dto"
	orderInterfaces "booker/modules/order/domain/interfaces"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/oaswrap/spec/adapter/fiberopenapi"
	"github.com/oaswrap/spec/option"
)

// RegisterRoutes sets up order HTTP routes.
func RegisterRoutes(r fiberopenapi.Router, orderSvc orderInterfaces.OrderService, tokenSvc interfaces.TokenService) {
	o := r.Group("/api/v1/orders", httpserver.AuthMiddleware(tokenSvc)).With(
		option.GroupSecurity("BearerAuth"),
		option.GroupTags("orders"),
	)

	o.Post("", CreateOrder(orderSvc)).With(
		option.OperationID("createOrder"),
		option.Summary("Create a new limit order"),
		option.Request(new(orderDTO.CreateOrderDTO)),
		option.Response(201, new(orderDTO.OrderResponse)),
	)
	o.Get("", ListOrders(orderSvc)).With(
		option.OperationID("listOrders"),
		option.Summary("List orders for current user"),
		option.Request(new(ListOrdersParam)),
		option.Response(200, new(orderDTO.OrderListResponse)),
	)
	o.Get("/:id", GetOrder(orderSvc)).With(
		option.OperationID("getOrder"),
		option.Summary("Get a single order by ID"),
		option.Request(new(OrderIDParam)),
		option.Response(200, new(orderDTO.OrderResponse)),
	)
	o.Delete("/:id", CancelOrder(orderSvc)).With(
		option.OperationID("cancelOrder"),
		option.Summary("Cancel an order"),
		option.Request(new(OrderIDParam)),
		option.Response(200, new(orderDTO.OrderResponse)),
	)
}
