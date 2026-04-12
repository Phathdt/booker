package interfaces

import (
	"context"

	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/entities"
)

type NotificationService interface {
	// CreateNotification persists and broadcasts. Returns true if new, false if duplicate.
	CreateNotification(ctx context.Context, n *entities.Notification) (bool, error)
	ListNotifications(ctx context.Context, userID string, req *dto.ListNotificationsDTO) ([]*entities.Notification, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) (int64, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
}
