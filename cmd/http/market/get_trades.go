package market

import (
	"booker/modules/market/trades"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetTrades godoc
// @Summary      Get recent trades for a pair
// @Tags         market
// @Produce      json
// @Param        pair   path   string  true   "Trading pair (e.g. BTC_USDT)"
// @Param        limit  query  int     false  "Number of trades (default 50, max 100)"
// @Success      200  {object}  httpserver.Response{data=[]TradeResponse}
// @Failure      404  {object}  httpserver.Response{error=object}
// @Router       /api/v1/market/trades/{pair} [get]
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
