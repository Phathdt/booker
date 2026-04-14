package users

import (
	"errors"

	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	apperrors "booker/pkg/errors"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RefreshToken godoc
func RefreshToken(cfg *config.Config, uc *usecases.RefreshTokenUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		refreshToken := c.Cookies(refreshTokenCookie)
		if refreshToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing refresh token")
		}

		result, err := uc.Execute(c.UserContext(), refreshToken)
		if err != nil {
			// Only clear cookie for auth-related errors (invalid/expired token, inactive user).
			// Transient errors (e.g. DB/Redis) should not invalidate the session.
			var appErr apperrors.AppError
			if errors.As(err, &appErr) {
				clearRefreshTokenCookie(c, cfg)
			}
			return err
		}

		setRefreshTokenCookie(c, cfg, result.RefreshToken)
		return httpserver.OK(c, dto.TokenPairResponse{
			AccessToken: result.AccessToken,
			ExpiresIn:   int(cfg.JWT.AccessTTL.Seconds()),
		})
	}
}
