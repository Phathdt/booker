package ws

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// UpgradeMiddleware checks if the request is a WebSocket upgrade.
func UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// Handler returns a Fiber handler for WebSocket connections.
func Handler(hub *Hub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := NewClient(c, hub)
		hub.Register(client)
		go client.WritePump()
		client.ReadPump()
	})
}
