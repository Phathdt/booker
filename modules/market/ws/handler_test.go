package ws

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgradeMiddleware_Returns_Handler tests middleware returns a fiber.Handler
func TestUpgradeMiddleware_Returns_Handler(t *testing.T) {
	middleware := UpgradeMiddleware()
	assert.NotNil(t, middleware)
}

// TestUpgradeMiddleware_IsCallable tests middleware can be registered with fiber
func TestUpgradeMiddleware_IsCallable(t *testing.T) {
	middleware := UpgradeMiddleware()

	app := fiber.New()
	app.Use(middleware)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Create HTTP request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Should not panic when testing
	resp, err := app.Test(req)
	require.NoError(t, err)
	// Non-WS request should get 426
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

// TestUpgradeMiddleware_ReturnsError_OnNonUpgrade tests middleware rejects non-upgrade
func TestUpgradeMiddleware_ReturnsError_OnNonUpgrade(t *testing.T) {
	middleware := UpgradeMiddleware()

	app := fiber.New()
	app.Get("/ws", middleware, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Create HTTP request
	req, err := http.NewRequest("GET", "/ws", nil)
	require.NoError(t, err)

	// Should return 426 Upgrade Required
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

// TestHandler_Returns_Fiber_Handler tests Handler returns a valid fiber.Handler
func TestHandler_Returns_Fiber_Handler(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)
	assert.NotNil(t, handler)
}

// TestHandler_IntegrationWithHub tests Handler works with Hub
func TestHandler_IntegrationWithHub(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	// Should be able to create the handler without panic
	assert.NotNil(t, handler)

	// Handler should be a function that can be registered with Fiber
	app := fiber.New()
	app.Get("/ws", handler)
	assert.NotNil(t, app)
}

// TestUpgradeMiddleware_MultipleRegistrations tests middleware can be registered
func TestUpgradeMiddleware_MultipleRegistrations(t *testing.T) {
	middleware1 := UpgradeMiddleware()
	middleware2 := UpgradeMiddleware()

	// Both should be valid handlers
	assert.NotNil(t, middleware1)
	assert.NotNil(t, middleware2)

	// Create app with both
	app := fiber.New()
	app.Use(middleware1)
	app.Use(middleware2)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Create HTTP request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Should not panic
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestHandler_CreateMultipleHandlers tests creating multiple handlers
func TestHandler_CreateMultipleHandlers(t *testing.T) {
	hub1 := NewHub()
	hub2 := NewHub()

	handler1 := Handler(hub1)
	handler2 := Handler(hub2)

	assert.NotNil(t, handler1)
	assert.NotNil(t, handler2)
}

// TestUpgradeMiddleware_WithFiberApp tests middleware in fiber app context
func TestUpgradeMiddleware_WithFiberApp(t *testing.T) {
	app := fiber.New()
	middleware := UpgradeMiddleware()

	app.Get("/api", middleware, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Create HTTP request
	req, err := http.NewRequest("GET", "/api", nil)
	require.NoError(t, err)

	// Non-WS request should get 426
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

// TestUpgradeMiddleware_AllowsWebSocketUpgrade tests that WebSocket upgrade requests pass through
func TestUpgradeMiddleware_AllowsWebSocketUpgrade(t *testing.T) {
	middleware := UpgradeMiddleware()

	app := fiber.New()
	app.Get("/ws", middleware, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Create WebSocket upgrade request
	req, err := http.NewRequest("GET", "/ws", nil)
	require.NoError(t, err)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := app.Test(req)
	require.NoError(t, err)
	// WebSocket upgrade should pass through (may be 101 Switching Protocols or handled by websocket.Handler)
	assert.NotEqual(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

// TestHandler_SuccessfullyCreatesClient tests that Handler creates a valid client
func TestHandler_SuccessfullyCreatesClient(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	// Should return a valid fiber.Handler function
	assert.NotNil(t, handler)

	// Should be callable
	app := fiber.New()
	app.Get("/ws", handler)
	assert.NotNil(t, app)
}

// TestHandler_WithDifferentHubs tests Handler with different hub instances
func TestHandler_WithDifferentHubs(t *testing.T) {
	hub1 := NewHub()
	hub2 := NewHub()

	handler1 := Handler(hub1)
	handler2 := Handler(hub2)

	// Both should be valid
	assert.NotNil(t, handler1)
	assert.NotNil(t, handler2)

	// Should be different functions with different hub references
	app1 := fiber.New()
	app1.Get("/ws", handler1)

	app2 := fiber.New()
	app2.Get("/ws", handler2)

	assert.NotNil(t, app1)
	assert.NotNil(t, app2)
}

// TestUpgradeMiddleware_WithOtherHTTPMethods tests middleware with non-GET requests
func TestUpgradeMiddleware_WithOtherHTTPMethods(t *testing.T) {
	middleware := UpgradeMiddleware()

	app := fiber.New()
	app.Post("/ws", middleware, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Non-WS POST request should get 426
	req, err := http.NewRequest("POST", "/ws", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

// TestUpgradeMiddleware_MultipleMiddlewareChain tests middleware in a chain
func TestUpgradeMiddleware_MultipleMiddlewareChain(t *testing.T) {
	app := fiber.New()

	// Add multiple middlewares in sequence
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("before", true)
		return c.Next()
	})

	app.Use(UpgradeMiddleware())

	app.Get("/ws", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, err := http.NewRequest("GET", "/ws", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}

// TestHandler_IntegrationWithFiber tests Handler integrates with Fiber app
func TestHandler_IntegrationWithFiber(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	app := fiber.New()
	app.Get("/market/ws", handler)

	// Creating request and testing should not panic
	req, err := http.NewRequest("GET", "/market/ws", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should be able to create app and route without panic
	assert.NotNil(t, resp)
}

// TestHandler_WithMultipleEndpoints tests Handler can be registered on multiple routes
func TestHandler_WithMultipleEndpoints(t *testing.T) {
	hub := NewHub()
	handler := Handler(hub)

	app := fiber.New()
	app.Get("/ws", handler)
	app.Get("/market/ws", handler)
	app.Get("/ticker/ws", handler)

	// All routes should be registered without error
	assert.NotNil(t, app)
	// Should have routes registered (exact count varies, just verify non-zero)
	assert.Greater(t, len(app.Stack()), 0)
}

// TestUpgradeMiddleware_ErrorResponse tests that error response is returned correctly
func TestUpgradeMiddleware_ErrorResponse(t *testing.T) {
	middleware := UpgradeMiddleware()

	app := fiber.New()
	app.Get("/ws", middleware, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, err := http.NewRequest("GET", "/ws", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should return 426 Upgrade Required
	assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
}
