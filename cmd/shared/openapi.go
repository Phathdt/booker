package shared

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oaswrap/spec/adapter/fiberopenapi"
	"github.com/oaswrap/spec/option"
)

// NewOpenAPIRouter creates a fiberopenapi Generator with common Booker API settings.
// Generator embeds Router, so it can be passed to RegisterRoutes functions.
func NewOpenAPIRouter(app *fiber.App) fiberopenapi.Generator {
	return fiberopenapi.NewRouter(app,
		option.WithTitle("Booker CEX API"),
		option.WithVersion("1.0"),
		option.WithDescription("Centralized Exchange demo — token trading platform"),
		option.WithSecurity("BearerAuth", option.SecurityHTTPBearer("Bearer")),
	)
}
