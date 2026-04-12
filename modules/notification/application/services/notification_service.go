package services

import (
	"context"
	"time"

	"booker/modules/notification/application/dto"
	"booker/modules/notification/domain/entities"
	"booker/modules/notification/domain/interfaces"
)

type notificationService struct {
	repo        interfaces.NotificationRepository
	broadcaster interfaces.Broadcaster
}

func NewNotificationService(
	repo interfaces.NotificationRepository,
	broadcaster interfaces.Broadcaster,
) interfaces.NotificationService {
	return &notificationService{repo: repo, broadcaster: broadcaster}
}

func (s *notificationService) CreateNotification(ctx context.Context, n *entities.Notification) (bool, error) {
	inserted, err := s.repo.Create(ctx, n)
	if err != nil {
		return false, err
	}
	// Only broadcast new notifications, skip duplicates
	if inserted && s.broadcaster != nil {
		s.broadcaster.SendToUser(n.UserID, n)
	}
	return inserted, nil
}

func (s *notificationService) ListNotifications(
	ctx context.Context,
	userID string,
	req *dto.ListNotificationsDTO,
) ([]*entities.Notification, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	cursor := time.Now()
	if req.Cursor != "" {
		parsed, err := time.Parse(time.RFC3339, req.Cursor)
		if err == nil {
			cursor = parsed
		}
	}

	return s.repo.ListByUser(ctx, userID, cursor, limit, req.OnlyUnread)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) (int64, error) {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) CountUnread(ctx context.Context, userID string) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}
