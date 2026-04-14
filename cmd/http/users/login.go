package users

import (
	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Login godoc
func Login(cfg *config.Config, uc *usecases.LoginUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.LoginDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		result, err := uc.Execute(c.UserContext(), req)
		if err != nil {
			return err
		}

		setRefreshTokenCookie(c, cfg, result.RefreshToken)
		return httpserver.OK(c, toAuthResponse(cfg, result.User, result.AccessToken))
	}
}
