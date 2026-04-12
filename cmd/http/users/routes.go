package users

import (
	"booker/config"
	"booker/modules/users/application/usecases"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes sets up all user/auth HTTP routes on the Fiber app.
func RegisterRoutes(
	app *fiber.App,
	cfg *config.Config,
	userSvc interfaces.UserService,
	tokenSvc interfaces.TokenService,
	registerUC *usecases.RegisterUseCase,
	loginUC *usecases.LoginUseCase,
	refreshTokenUC *usecases.RefreshTokenUseCase,
	logoutUC *usecases.LogoutUseCase,
) {
	api := app.Group("/api/v1")

	// Public auth routes
	auth := api.Group("/auth")
	auth.Post("/register", Register(cfg, registerUC))
	auth.Post("/login", Login(cfg, loginUC))
	auth.Post("/refresh", RefreshToken(cfg, refreshTokenUC))

	// Protected auth routes
	authProtected := auth.Group("", httpserver.AuthMiddleware(tokenSvc))
	authProtected.Post("/logout", Logout(cfg, logoutUC))
	authProtected.Get("/me", GetMe(userSvc))

	// Protected user routes
	usersGroup := api.Group("/users", httpserver.AuthMiddleware(tokenSvc))
	usersGroup.Get("/:id", GetUser(userSvc))
	usersGroup.Get("/", ListUsers(userSvc))
}
