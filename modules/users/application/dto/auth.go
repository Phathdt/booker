package dto

// RegisterDTO holds registration input.
type RegisterDTO struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// LoginDTO holds login input.
type LoginDTO struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenDTO is no longer used — refresh token is read from HTTP-only cookie.
