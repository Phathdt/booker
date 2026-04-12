package users

import (
	"time"

	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RefreshToken godoc
// @Summary      Refresh access token using HTTP-only cookie
// @Tags         auth
// @Produce      json
// @Success      200   {object}  httpserver.Response{data=dto.TokenPairResponse}
// @Failure      401   {object}  httpserver.Response{error=object}
// @Router       /api/v1/auth/refresh [post]
func RefreshToken(cfg *config.Config, uc *usecases.RefreshTokenUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		refreshToken := c.Cookies(refreshTokenCookie)
		if refreshToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing refresh token")
		}

		result, err := uc.Execute(c.UserContext(), refreshToken)
		if err != nil {
			clearRefreshTokenCookie(c, cfg)
			return err
		}

		setRefreshTokenCookie(c, cfg, result.RefreshToken)
		return httpserver.OK(c, dto.TokenPairResponse{
			AccessToken: result.AccessToken,
			ExpiresIn:   int(15 * time.Minute / time.Second),
		})
	}
}
