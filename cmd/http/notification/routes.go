package notification

import (
	"booker/modules/notification/domain/interfaces"
	"booker/modules/notification/infrastructure/ws"
	userInterfaces "booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes sets up notification HTTP + WebSocket routes.
func RegisterRoutes(
	app *fiber.App,
	svc interfaces.NotificationService,
	tokenSvc userInterfaces.TokenService,
	hub *ws.Hub,
) {
	// REST endpoints (JWT auth via header)
	n := app.Group("/api/v1/notifications", httpserver.AuthMiddleware(tokenSvc))
	n.Get("/", ListNotifications(svc))
	n.Patch("/:id/read", MarkAsRead(svc))
	n.Post("/read-all", MarkAllAsRead(svc))
	n.Get("/unread-count", UnreadCount(svc))

	// WebSocket endpoint (JWT auth via query param or header)
	app.Use("/api/v1/notifications/ws", WSUpgradeMiddleware(tokenSvc))
	app.Get("/api/v1/notifications/ws", WSHandler(hub))
}
