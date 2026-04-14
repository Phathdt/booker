package users

// UserIDParam documents the user ID path parameter.
type UserIDParam struct {
	ID string `params:"id" required:"true"`
}

// ListUsersParam documents the list users query parameters.
type ListUsersParam struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}
