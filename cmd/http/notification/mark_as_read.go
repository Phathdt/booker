package notification

import (
	"booker/modules/notification/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// MarkAsRead marks a single notification as read.
func MarkAsRead(svc interfaces.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		id := c.Params("id")

		if err := svc.MarkAsRead(c.UserContext(), id, userID); err != nil {
			return err
		}

		return httpserver.OK(c, fiber.Map{"message": "marked as read"})
	}
}
