package users

import (
	"time"

	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/domain/entities"
)

func toAuthResponse(cfg *config.Config, u *entities.User, access string) dto.AuthResponse {
	return dto.AuthResponse{
		User:        toUserResponse(u),
		AccessToken: access,
		ExpiresIn:   int(cfg.JWT.AccessTTL.Seconds()),
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
