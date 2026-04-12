package notification

import (
	"strings"

	"booker/modules/notification/infrastructure/ws"
	userInterfaces "booker/modules/users/domain/interfaces"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// WSUpgradeMiddleware validates JWT from query param before upgrading to WebSocket.
func WSUpgradeMiddleware(tokenSvc userInterfaces.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		// Token from query param: ?token=xxx
		token := c.Query("token")
		if token == "" {
			// Fallback: Authorization header
			auth := c.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				token = auth[7:]
			}
		}
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing authentication token")
		}

		claims, err := tokenSvc.ValidateAccessToken(c.UserContext(), token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}

// WSHandler handles WebSocket connections for real-time notifications.
func WSHandler(hub *ws.Hub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			_ = c.Close()
			return
		}
		hub.Register(userID, c)
		defer hub.Unregister(userID, c)

		// Keep connection alive — read loop (handles pings/close frames)
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	})
}
