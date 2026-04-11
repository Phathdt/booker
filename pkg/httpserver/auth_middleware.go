package httpserver

import (
	"strings"

	"booker/modules/users/domain/interfaces"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT and injects user_id + role into Locals.
func AuthMiddleware(tokenSvc interfaces.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing or invalid Authorization header")
		}

		token := authHeader[7:]
		claims, err := tokenSvc.ValidateAccessToken(c.UserContext(), token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}
