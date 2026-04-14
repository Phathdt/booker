package dto

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID        string `json:"id"         required:"true" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email"      required:"true" example:"user@example.com"`
	Role      string `json:"role"       required:"true" example:"user"`
	Status    string `json:"status"     required:"true" example:"active"`
	CreatedAt string `json:"created_at" required:"true" example:"2026-04-12T00:00:00Z"`
	UpdatedAt string `json:"updated_at" required:"true" example:"2026-04-12T00:00:00Z"`
}

// AuthResponse represents the response for register/login.
// Refresh token is delivered via HTTP-only cookie, not in the response body.
type AuthResponse struct {
	User        UserResponse `json:"user"         required:"true"`
	AccessToken string       `json:"access_token" required:"true" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn   int          `json:"expires_in"   required:"true" example:"900"`
}

// TokenPairResponse represents the response for token refresh.
// Refresh token is delivered via HTTP-only cookie, not in the response body.
type TokenPairResponse struct {
	AccessToken string `json:"access_token" required:"true" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn   int    `json:"expires_in"   required:"true" example:"900"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message" required:"true" example:"logged out"`
}

// UserListResponse represents a paginated list of users.
type UserListResponse struct {
	Users []UserResponse `json:"users" required:"true"`
	Total int64          `json:"total" required:"true" example:"42"`
}
