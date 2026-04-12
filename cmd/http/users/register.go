package users

import (
	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterDTO  true  "Register request"
// @Success      201   {object}  httpserver.Response{data=dto.AuthResponse}
// @Failure      400   {object}  httpserver.Response{error=object}
// @Failure      409   {object}  httpserver.Response{error=object}
// @Router       /api/v1/auth/register [post]
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
