package notification

import (
	notifDTO "booker/modules/notification/application/dto"
	"booker/modules/notification/domain/interfaces"
	"booker/modules/notification/infrastructure/ws"
	userInterfaces "booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
	"github.com/oaswrap/spec/adapter/fiberopenapi"
	"github.com/oaswrap/spec/option"
)

// RegisterRoutes sets up notification HTTP + WebSocket routes.
func RegisterRoutes(
	app *fiber.App,
	r fiberopenapi.Router,
	svc interfaces.NotificationService,
	tokenSvc userInterfaces.TokenService,
	hub *ws.Hub,
) {
	// REST endpoints (JWT auth via header)
	n := r.Group("/api/v1/notifications", httpserver.AuthMiddleware(tokenSvc)).With(
		option.GroupSecurity("BearerAuth"),
		option.GroupTags("notifications"),
	)

	n.Get("", ListNotifications(svc)).With(
		option.OperationID("listNotifications"),
		option.Summary("List notifications for current user"),
		option.Request(new(ListNotificationsParam)),
		option.Response(200, new(notifDTO.NotificationListResponse)),
	)
	n.Patch("/:id/read", MarkAsRead(svc)).With(
		option.OperationID("markNotificationAsRead"),
		option.Summary("Mark a notification as read"),
		option.Request(new(NotificationIDParam)),
		option.Response(200, new(fiber.Map)),
	)
	n.Post("/read-all", MarkAllAsRead(svc)).With(
		option.OperationID("markAllNotificationsAsRead"),
		option.Summary("Mark all notifications as read"),
		option.Response(200, new(fiber.Map)),
	)
	n.Get("/unread-count", UnreadCount(svc)).With(
		option.OperationID("getUnreadCount"),
		option.Summary("Get unread notification count"),
		option.Response(200, new(notifDTO.UnreadCountResponse)),
	)

	// WebSocket endpoint (not documented in OpenAPI)
	app.Use("/api/v1/notifications/ws", WSUpgradeMiddleware(tokenSvc))
	app.Get("/api/v1/notifications/ws", WSHandler(hub))
}
