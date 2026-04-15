package market

import (
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetPairs godoc
func GetPairs(pairs []PairInfo) fiber.Handler {
	// Pre-compute response
	resp := make([]PairResponse, len(pairs))
	for i, p := range pairs {
		resp[i] = PairResponse{
			ID:         p.ID,
			BaseAsset:  p.BaseAsset,
			QuoteAsset: p.QuoteAsset,
			MinQty:     p.MinQty,
			TickSize:   p.TickSize,
		}
	}

	return func(c *fiber.Ctx) error {
		return httpserver.OK(c, resp)
	}
}
