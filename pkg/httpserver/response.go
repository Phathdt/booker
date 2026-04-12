package httpserver

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

// Response is the standard API response format.
type Response struct {
	Data      any    `json:"data,omitempty"`
	Error     any    `json:"error,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// OK sends a 200 response with data + trace context.
func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(buildResponse(c, data, nil))
}

// Created sends a 201 response.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(buildResponse(c, data, nil))
}

// ErrorResponse sends an error response.
func ErrorResponse(c *fiber.Ctx, status int, err any) error {
	return c.Status(status).JSON(buildResponse(c, nil, err))
}

func buildResponse(c *fiber.Ctx, data, errData any) Response {
	resp := Response{
		Data:      data,
		Error:     errData,
		RequestID: localsString(c, "request_id"),
	}

	// Extract trace_id from OTel span context
	span := trace.SpanFromContext(c.UserContext())
	if span.SpanContext().HasTraceID() {
		resp.TraceID = span.SpanContext().TraceID().String()
	}

	return resp
}

func localsString(c *fiber.Ctx, key string) string {
	v, _ := c.Locals(key).(string)
	return v
}
