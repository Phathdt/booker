package market

import (
	"booker/modules/market/ticker"
	"booker/modules/market/trades"
	"booker/modules/market/ws"
	pb "booker/proto/matching/v1/gen"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes sets up market HTTP + WS routes on the Fiber app.
func RegisterRoutes(
	app *fiber.App,
	tickers map[string]*ticker.Aggregator,
	recentTrades map[string]*trades.RecentTrades,
	pairs []PairInfo,
	hub *ws.Hub,
	matchingClient pb.MatchingServiceClient,
) {
	m := app.Group("/api/v1/market")

	m.Get("/pairs", GetPairs(pairs))
	m.Get("/ticker", GetTickers(tickers))
	m.Get("/ticker/:pair", GetTicker(tickers))
	m.Get("/trades/:pair", GetTrades(recentTrades))
	m.Get("/orderbook/:pair", GetOrderBook(matchingClient))

	// WebSocket
	app.Use("/ws", ws.UpgradeMiddleware())
	app.Get("/ws", ws.Handler(hub))
}
