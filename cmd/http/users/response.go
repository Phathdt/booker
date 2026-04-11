package users

import (
	"time"

	"booker/modules/users/application/dto"
	"booker/modules/users/domain/entities"
)

func toAuthResponse(u *entities.User, access, refresh string) dto.AuthResponse {
	return dto.AuthResponse{
		User:         toUserResponse(u),
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(15 * time.Minute / time.Second),
	}
}

func toUserResponse(u *entities.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}
