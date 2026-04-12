package notification

import (
	"booker/modules/notification/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// MarkAllAsRead marks all notifications as read for the authenticated user.
func MarkAllAsRead(svc interfaces.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}

		count, err := svc.MarkAllAsRead(c.UserContext(), userID)
		if err != nil {
			return err
		}

		return httpserver.OK(c, fiber.Map{"marked": count})
	}
}
