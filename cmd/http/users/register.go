package users

import (
	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Register godoc
func Register(cfg *config.Config, uc *usecases.RegisterUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req dto.RegisterDTO
		if err := httpserver.BindAndValidate(c, &req); err != nil {
			return err
		}

		result, err := uc.Execute(c.UserContext(), req)
		if err != nil {
			return err
		}

		setRefreshTokenCookie(c, cfg, result.RefreshToken)
		return httpserver.Created(c, toAuthResponse(cfg, result.User, result.AccessToken))
	}
}
