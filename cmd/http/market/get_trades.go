package market

import (
	"booker/modules/market/trades"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetTrades godoc
func GetTrades(recentTrades map[string]*trades.RecentTrades) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pair := c.Params("pair")
		rt, ok := recentTrades[pair]
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, "Unknown trading pair")
		}

		limit := c.QueryInt("limit", 50)
		if limit > 100 {
			limit = 100
		}

		tradeList := rt.GetRecent(limit)
		result := make([]TradeResponse, len(tradeList))
		for i, t := range tradeList {
			result[i] = TradeResponse{
				TradeID:   t.TradeID,
				Price:     t.Price,
				Quantity:  t.Quantity,
				Timestamp: t.Timestamp,
			}
		}

		return httpserver.OK(c, result)
	}
}
