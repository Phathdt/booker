package users

import (
	"booker/config"
	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// Login godoc
// @Summary      Login with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginDTO  true  "Login request"
// @Success      200   {object}  httpserver.Response{data=dto.AuthResponse}
// @Failure      400   {object}  httpserver.Response{error=object}
// @Failure      401   {object}  httpserver.Response{error=object}
// @Router       /api/v1/auth/login [post]
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
