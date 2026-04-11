package users

import (
	"time"

	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshTokenDTO  true  "Refresh token request"
// @Success      200   {object}  httpserver.Response{data=dto.TokenPairResponse}
// @Failure      400   {object}  httpserver.Response{error=object}
// @Failure      401   {object}  httpserver.Response{error=object}
// @Router       /api/v1/auth/refresh [post]
func RefreshToken(uc *usecases.RefreshTokenUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.RefreshTokenDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		result, err := uc.Execute(c.UserContext(), req.RefreshToken)
		if err != nil {
			return err
		}

		return httpserver.OK(c, dto.TokenPairResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresIn:    int(15 * time.Minute / time.Second),
		})
	}
}
