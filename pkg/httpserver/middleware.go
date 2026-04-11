package httpserver

import (
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

// TracingMiddleware creates an OTel span for each HTTP request.
func TracingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), c.Method()+" "+c.Path(),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.request_id", c.Locals("request_id", "").(string)),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)
		err := c.Next()

		span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))
		return err
	}
}
