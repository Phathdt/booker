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

// RefreshTokenDTO holds refresh token input.
type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
