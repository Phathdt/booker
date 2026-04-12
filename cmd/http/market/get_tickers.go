package market

import (
	"booker/modules/market/ticker"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// GetTickers godoc
// @Summary      Get all pair tickers
// @Tags         market
// @Produce      json
// @Success      200  {object}  httpserver.Response{data=[]TickerResponse}
// @Router       /api/v1/market/ticker [get]
func GetTickers(tickers map[string]*ticker.Aggregator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := make([]TickerResponse, 0, len(tickers))
		for pair, agg := range tickers {
			t := agg.GetTicker()
			result = append(result, TickerResponse{
				Pair:      pair,
				Open:      t.Open.String(),
				High:      t.High.String(),
				Low:       t.Low.String(),
				Close:     t.Close.String(),
				Volume:    t.Volume.String(),
				ChangePct: t.ChangePct.String(),
				LastPrice: t.LastPrice.String(),
				Timestamp: t.Timestamp,
			})
		}
		return httpserver.OK(c, result)
	}
}

// GetTicker godoc
// @Summary      Get ticker for a single pair
// @Tags         market
// @Produce      json
// @Param        pair  path  string  true  "Trading pair (e.g. BTC_USDT)"
// @Success      200  {object}  httpserver.Response{data=TickerResponse}
// @Failure      404  {object}  httpserver.Response{error=object}
// @Router       /api/v1/market/ticker/{pair} [get]
func GetTicker(tickers map[string]*ticker.Aggregator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pair := c.Params("pair")
		agg, ok := tickers[pair]
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, "Unknown trading pair")
		}

		t := agg.GetTicker()
		return httpserver.OK(c, TickerResponse{
			Pair:      pair,
			Open:      t.Open.String(),
			High:      t.High.String(),
			Low:       t.Low.String(),
			Close:     t.Close.String(),
			Volume:    t.Volume.String(),
			ChangePct: t.ChangePct.String(),
			LastPrice: t.LastPrice.String(),
			Timestamp: t.Timestamp,
		})
	}
}
