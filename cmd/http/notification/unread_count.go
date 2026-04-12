package notification

import (
	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// UnreadCount returns the count of unread notifications.
func UnreadCount(svc interfaces.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		count, err := svc.CountUnread(c.UserContext(), userID)
		if err != nil {
			return err
		}

		return httpserver.OK(c, dto.UnreadCountResponse{Count: count})
	}
}
