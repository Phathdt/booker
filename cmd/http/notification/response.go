package notification

import (
	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/entities"
)

func toNotificationResponse(n *entities.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:        n.ID,
		Type:      string(n.Type),
		Title:     n.Title,
		Body:      n.Body,
		Metadata:  n.Metadata,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toNotificationListResponse(notifs []*entities.Notification) dto.NotificationListResponse {
	items := make([]dto.NotificationResponse, len(notifs))
	for i, n := range notifs {
		items[i] = toNotificationResponse(n)
	}
	return dto.NotificationListResponse{Notifications: items}
}
