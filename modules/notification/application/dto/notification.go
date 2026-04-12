package dto

type ListNotificationsDTO struct {
	Cursor     string `query:"cursor"      validate:"omitempty"`
	Limit      int32  `query:"limit"       validate:"omitempty,min=1,max=50"`
	OnlyUnread bool   `query:"only_unread"`
}

type NotificationResponse struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Metadata  map[string]string `json:"metadata"`
	IsRead    bool              `json:"is_read"`
	CreatedAt string            `json:"created_at"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
