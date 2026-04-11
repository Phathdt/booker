package entities

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBanned   = "banned"
)

// User represents a domain user entity.
type User struct {
	ID        string
	Email     string
	Password  string
	Role      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccessClaims holds JWT access token claims.
type AccessClaims struct {
	UserID string
	Email  string
	Role   string
	JTI    string
}

// RefreshClaims holds JWT refresh token claims.
type RefreshClaims struct {
	UserID string
	JTI    string
}
