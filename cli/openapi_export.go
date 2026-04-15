package cli

import (
	"fmt"

	marketHTTP "booker/cmd/http/market"
	notifHTTP "booker/cmd/http/notification"
	orderHTTP "booker/cmd/http/order"
	usersHTTP "booker/cmd/http/users"
	walletHTTP "booker/cmd/http/wallet"
	"booker/cmd/shared"

	"github.com/gofiber/fiber/v2"
	urfavecli "github.com/urfave/cli/v2"
)

// RunOpenAPIExport generates a combined OpenAPI spec file without starting servers.
// It registers all routes on a dummy Fiber app and writes the spec to disk.
func RunOpenAPIExport(c *urfavecli.Context) error {
	output := c.String("output")
	if output == "" {
		output = "docs/openapi.yaml"
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	r := shared.NewOpenAPIRouter(app)

	// Register all module routes with nil dependencies (handlers are never called).
	// fiberopenapi only needs the route metadata (.With() chains), not working handlers.
	usersHTTP.RegisterRoutes(r, nil, nil, nil, nil, nil, nil, nil)
	walletHTTP.RegisterRoutes(r, nil, nil)
	orderHTTP.RegisterRoutes(r, nil, nil)
	marketHTTP.RegisterRoutes(app, r, nil, nil, nil, nil, nil)
	notifHTTP.RegisterRoutes(app, r, nil, nil, nil)

	if err := r.WriteSchemaTo(output); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec: %w", err)
	}

	fmt.Printf("OpenAPI spec written to %s\n", output)
	return nil
}
