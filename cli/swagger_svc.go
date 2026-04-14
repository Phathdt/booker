package cli

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"booker/cmd/shared"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	urfavecli "github.com/urfave/cli/v2"
)

// RunSwaggerSvc starts a lightweight HTTP server that serves OpenAPI docs
// via the fiberopenapi built-in /docs UI.
func RunSwaggerSvc(c *urfavecli.Context) error {
	httpPort := c.Int("http-port")

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := fiber.New(fiber.Config{
		AppName:               "booker-swagger-svc",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(cors.New())

	// Initialize OpenAPI router — serves /docs and /docs/openapi.yaml
	_ = shared.NewOpenAPIRouter(app)

	// Health check for Traefik / Docker
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	httpAddr := fmt.Sprintf(":%d", httpPort)
	errCh := make(chan error, 1)
	go func() {
		log.Info("OpenAPI docs started", "address", httpAddr)
		if err := app.Listen(httpAddr); err != nil {
			errCh <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("swagger server failed: %w", err)
	case <-sigChan:
		log.Info("shutting down swagger service...")
		return app.Shutdown()
	}
}
