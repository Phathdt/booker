package httpserver

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("httpserver")

// RequestIDMiddleware generates a unique request ID for each request.
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Locals("request_id", requestID)
		c.Set("X-Request-Id", requestID)
		return c.Next()
	}
}

// LoggingMiddleware logs each HTTP request with method, path, status, and latency.
func LoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		status := c.Response().StatusCode()
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		reqID, _ := c.Locals("request_id").(string)
		slog.LogAttrs(c.UserContext(), level, "HTTP request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.String("latency", latency.String()),
			slog.String("request_id", reqID),
		)
		return err
	}
}

// LogRoutes logs all registered routes on the Fiber app at startup.
func LogRoutes(app *fiber.App, serviceName string) {
	for _, routes := range app.Stack() {
		for _, route := range routes {
			if route.Path == "/" || route.Method == "HEAD" {
				continue
			}
			slog.Info("route registered",
				slog.String("service", serviceName),
				slog.String("method", route.Method),
				slog.String("path", route.Path),
			)
		}
	}
}

// TracingMiddleware creates an OTel span for each HTTP request.
func TracingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), c.Method()+" "+c.Path(),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.request_id", localsString(c, "request_id")),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)
		err := c.Next()

		span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))
		return err
	}
}
