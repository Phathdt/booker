package repositories

import (
	"context"
	"encoding/json"
	"time"

	"booker/modules/notification/domain"
	"booker/modules/notification/domain/entities"
	"booker/modules/notification/domain/interfaces"
	"booker/modules/notification/infrastructure/gen"

	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationRepository struct {
	q *gen.Queries
}

func NewNotificationRepository(pool *pgxpool.Pool) interfaces.NotificationRepository {
	return &notificationRepository{q: gen.New(pool)}
}

func (r *notificationRepository) Create(ctx context.Context, n *entities.Notification) (bool, error) {
	metadata, err := json.Marshal(n.Metadata)
	if err != nil {
		return false, err
	}
	rows, err := r.q.CreateNotification(ctx, gen.CreateNotificationParams{
		UserID:   n.UserID,
		EventKey: n.EventKey,
		Type:     string(n.Type),
		Title:    n.Title,
		Body:     n.Body,
		Metadata: metadata,
	})
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID string, cursor time.Time, limit int32, onlyUnread bool) ([]*entities.Notification, error) {
	if onlyUnread {
		rows, err := r.q.ListUnreadNotificationsByUser(ctx, gen.ListUnreadNotificationsByUserParams{
			UserID:    userID,
			CreatedAt: cursor,
			Limit:     limit,
		})
		if err != nil {
			return nil, err
		}
		return toEntities(rows), nil
	}

	rows, err := r.q.ListNotificationsByUser(ctx, gen.ListNotificationsByUserParams{
		UserID:    userID,
		CreatedAt: cursor,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	return toEntities(rows), nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID string) error {
	affected, err := r.q.MarkNotificationAsRead(ctx, gen.MarkNotificationAsReadParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotificationNotFound
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) (int64, error) {
	return r.q.MarkAllNotificationsAsRead(ctx, userID)
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	return r.q.CountUnreadNotifications(ctx, userID)
}

func toEntities(rows []gen.Notification) []*entities.Notification {
	result := make([]*entities.Notification, len(rows))
	for i, row := range rows {
		result[i] = toEntity(row)
	}
	return result
}

func toEntity(row gen.Notification) *entities.Notification {
	metadata := make(map[string]string)
	_ = json.Unmarshal(row.Metadata, &metadata)

	return &entities.Notification{
		ID:        row.ID,
		UserID:    row.UserID,
		EventKey:  row.EventKey,
		Type:      entities.NotificationType(row.Type),
		Title:     row.Title,
		Body:      row.Body,
		Metadata:  metadata,
		IsRead:    row.IsRead,
		CreatedAt: row.CreatedAt,
	}
}
