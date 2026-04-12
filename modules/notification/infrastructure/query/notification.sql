-- name: CreateNotification :execrows
INSERT INTO notifications (user_id, event_key, type, title, body, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (event_key) DO NOTHING;

-- name: ListNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;

-- name: ListUnreadNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1 AND is_read = false AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;

-- name: MarkNotificationAsRead :execrows
UPDATE notifications SET is_read = true
WHERE id = $1 AND user_id = $2 AND is_read = false;

-- name: MarkAllNotificationsAsRead :execrows
UPDATE notifications SET is_read = true
WHERE user_id = $1 AND is_read = false;

-- name: CountUnreadNotifications :one
SELECT count(*) FROM notifications
WHERE user_id = $1 AND is_read = false;
