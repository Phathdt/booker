package interfaces

import (
	"context"
	"time"

	"booker/modules/notification/domain/entities"
)

type NotificationRepository interface {
	// Create inserts a notification. Returns true if inserted, false if duplicate (dedup by event_key).
	Create(ctx context.Context, n *entities.Notification) (bool, error)
	ListByUser(ctx context.Context, userID string, cursor time.Time, limit int32, onlyUnread bool) ([]*entities.Notification, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) (int64, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
}
