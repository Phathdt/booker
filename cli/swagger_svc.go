package cli

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "booker/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/swaggo/swag"
	urfavecli "github.com/urfave/cli/v2"
)

const scalarHTML = `<!doctype html>
<html>
<head>
    <title>Booker CEX API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
    <script id="api-reference" data-url="/swagger/doc.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

// RunSwaggerSvc starts a lightweight HTTP server that serves Scalar API docs.
func RunSwaggerSvc(c *urfavecli.Context) error {
	httpPort := c.Int("http-port")

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := fiber.New(fiber.Config{
		AppName:               "booker-swagger-svc",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(cors.New())

	// Serve OpenAPI spec JSON
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		doc, err := swag.ReadDoc()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		c.Set("Content-Type", "application/json")
		return c.SendString(doc)
	})

	// Serve Scalar UI
	app.Get("/swagger", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(scalarHTML)
	})

	// Health check for Traefik / Docker
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	httpAddr := fmt.Sprintf(":%d", httpPort)
	errCh := make(chan error, 1)
	go func() {
		log.Info("Scalar API docs started", "address", httpAddr)
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
