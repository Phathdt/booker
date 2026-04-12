package dto

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID        string `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email"      example:"user@example.com"`
	Role      string `json:"role"       example:"user"`
	Status    string `json:"status"     example:"active"`
	CreatedAt string `json:"created_at" example:"2026-04-12T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-04-12T00:00:00Z"`
}

// AuthResponse represents the response for register/login.
// Refresh token is delivered via HTTP-only cookie, not in the response body.
type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn   int          `json:"expires_in"   example:"900"`
}

// TokenPairResponse represents the response for token refresh.
// Refresh token is delivered via HTTP-only cookie, not in the response body.
type TokenPairResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn   int    `json:"expires_in"   example:"900"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message" example:"logged out"`
}

// UserListResponse represents a paginated list of users.
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total" example:"42"`
}
