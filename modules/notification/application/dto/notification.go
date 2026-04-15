package dto

type ListNotificationsDTO struct {
	Cursor     string `query:"cursor"      validate:"omitempty"`
	Limit      int32  `query:"limit"       validate:"omitempty,min=1,max=50"`
	OnlyUnread bool   `query:"onlyUnread"`
}

type NotificationResponse struct {
	ID        string            `json:"id"         required:"true" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type      string            `json:"type"       required:"true" example:"order_filled"`
	Title     string            `json:"title"      required:"true" example:"Order Filled"`
	Body      string            `json:"body"       required:"true" example:"Your buy order for 0.5 BTC was filled"`
	Metadata  map[string]string `json:"metadata"   required:"true"`
	IsRead    bool              `json:"isRead"    required:"true" example:"false"`
	CreatedAt string            `json:"createdAt" required:"true" example:"2026-04-12T00:00:00Z"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications" required:"true"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count" required:"true" example:"5"`
}
