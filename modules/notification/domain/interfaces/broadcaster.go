package interfaces

import "booker/modules/notification/domain/entities"

// Broadcaster pushes notifications to connected WebSocket clients.
type Broadcaster interface {
	SendToUser(userID string, notification *entities.Notification)
}
