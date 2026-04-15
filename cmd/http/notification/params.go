package notification

// NotificationIDParam documents the notification ID path parameter.
type NotificationIDParam struct {
	ID string `params:"id" required:"true"`
}

// ListNotificationsParam documents the list notifications query parameters.
type ListNotificationsParam struct {
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
	OnlyUnread bool   `query:"onlyUnread"`
}
