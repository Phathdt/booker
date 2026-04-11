package gateway

import (
	"net/http"

	"booker/pkg/logger"

	"github.com/rs/cors"
)

// NewHTTPServer creates an HTTP server with CORS and response wrapping middleware.
func NewHTTPServer(addr string, gwMux http.Handler, log logger.Logger) *http.Server {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(WrapResponse(gwMux))

	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
