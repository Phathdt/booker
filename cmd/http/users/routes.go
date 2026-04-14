package users

import (
	"booker/config"
	userDTO "booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/oaswrap/spec/adapter/fiberopenapi"
	"github.com/oaswrap/spec/option"
)

// RegisterRoutes sets up all user/auth HTTP routes.
func RegisterRoutes(
	r fiberopenapi.Router,
	cfg *config.Config,
	userSvc interfaces.UserService,
	tokenSvc interfaces.TokenService,
	registerUC *usecases.RegisterUseCase,
	loginUC *usecases.LoginUseCase,
	refreshTokenUC *usecases.RefreshTokenUseCase,
	logoutUC *usecases.LogoutUseCase,
) {
	api := r.Group("/api/v1")

	// Public auth routes
	auth := api.Group("/auth").With(option.GroupTags("auth"))
	auth.Post("/register", Register(cfg, registerUC)).With(
		option.Summary("Register a new user"),
		option.Request(new(userDTO.RegisterDTO)),
		option.Response(201, new(userDTO.AuthResponse)),
	)
	auth.Post("/login", Login(cfg, loginUC)).With(
		option.Summary("Login with email and password"),
		option.Request(new(userDTO.LoginDTO)),
		option.Response(200, new(userDTO.AuthResponse)),
	)
	auth.Post("/refresh", RefreshToken(cfg, refreshTokenUC)).With(
		option.Summary("Refresh access token"),
		option.Response(200, new(userDTO.TokenPairResponse)),
	)

	// Protected auth routes
	authProtected := auth.Group("", httpserver.AuthMiddleware(tokenSvc)).With(
		option.GroupSecurity("BearerAuth"),
	)
	authProtected.Post("/logout", Logout(cfg, logoutUC)).With(
		option.Summary("Logout (revoke all tokens)"),
		option.Response(200, new(userDTO.MessageResponse)),
	)
	authProtected.Get("/me", GetMe(userSvc)).With(
		option.Summary("Get current authenticated user"),
		option.Response(200, new(userDTO.UserResponse)),
	)

	// Protected user routes
	usersGroup := api.Group("/users", httpserver.AuthMiddleware(tokenSvc)).With(
		option.GroupSecurity("BearerAuth"),
		option.GroupTags("users"),
	)
	usersGroup.Get("/:id", GetUser(userSvc)).With(
		option.Summary("Get user by ID"),
		option.Request(new(UserIDParam)),
		option.Response(200, new(userDTO.UserResponse)),
	)
	usersGroup.Get("/", ListUsers(userSvc)).With(
		option.Summary("List users"),
		option.Request(new(ListUsersParam)),
		option.Response(200, new(userDTO.UserListResponse)),
	)
}
