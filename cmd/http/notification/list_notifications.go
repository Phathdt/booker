package notification

import (
	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// ListNotifications returns paginated notifications for the authenticated user.
func ListNotifications(svc interfaces.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.ListNotificationsDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		userID := c.Locals("user_id").(string)
		notifs, err := svc.ListNotifications(c.UserContext(), userID, &req)
		if err != nil {
			return err
		}

		return httpserver.OK(c, toNotificationListResponse(notifs))
	}
}
