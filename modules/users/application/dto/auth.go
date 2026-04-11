package dto

// RegisterDTO holds registration input.
type RegisterDTO struct {
	Email    string
	Password string
}

// LoginDTO holds login input.
type LoginDTO struct {
	Email    string
	Password string
}
