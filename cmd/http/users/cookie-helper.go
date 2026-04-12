package users

import (
	"time"

	"booker/config"

	"github.com/gofiber/fiber/v2"
)

const refreshTokenCookie = "refresh_token"

func setRefreshTokenCookie(c *fiber.Ctx, cfg *config.Config, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   cfg.Env == "production",
		MaxAge:   int(cfg.JWT.RefreshTTL / time.Second),
		Path:     "/",
	})
}

func clearRefreshTokenCookie(c *fiber.Ctx, cfg *config.Config) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   cfg.Env == "production",
		MaxAge:   -1,
		Path:     "/",
	})
}
