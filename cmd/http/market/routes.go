package market

import (
	"booker/modules/market/ticker"
	"booker/modules/market/trades"
	"booker/modules/market/ws"
	pb "booker/proto/matching/v1/gen"

	"github.com/gofiber/fiber/v2"
	"github.com/oaswrap/spec/adapter/fiberopenapi"
	"github.com/oaswrap/spec/option"
)

// RegisterRoutes sets up market HTTP + WS routes on the Fiber app.
func RegisterRoutes(
	app *fiber.App,
	r fiberopenapi.Router,
	tickers map[string]*ticker.Aggregator,
	recentTrades map[string]*trades.RecentTrades,
	pairs []PairInfo,
	hub *ws.Hub,
	matchingClient pb.MatchingServiceClient,
) {
	m := r.Group("/api/v1/market")

	m.Get("/pairs", GetPairs(pairs)).With(
		option.OperationID("getPairs"),
		option.Summary("List active trading pairs"),
		option.Tags("market"),
		option.Response(200, new([]PairResponse)),
	)
	m.Get("/ticker", GetTickers(tickers)).With(
		option.OperationID("getTickers"),
		option.Summary("Get all pair tickers"),
		option.Tags("market"),
		option.Response(200, new([]TickerResponse)),
	)
	m.Get("/ticker/:pair", GetTicker(tickers)).With(
		option.OperationID("getTicker"),
		option.Summary("Get ticker for a single pair"),
		option.Tags("market"),
		option.Request(new(PairPathParam)),
		option.Response(200, new(TickerResponse)),
	)
	m.Get("/trades/:pair", GetTrades(recentTrades)).With(
		option.OperationID("getTrades"),
		option.Summary("Get recent trades for a pair"),
		option.Tags("market"),
		option.Request(new(TradesQueryParam)),
		option.Response(200, new([]TradeResponse)),
	)
	m.Get("/orderbook/:pair", GetOrderBook(matchingClient)).With(
		option.OperationID("getOrderBook"),
		option.Summary("Get order book depth for a trading pair"),
		option.Tags("market"),
		option.Request(new(OrderBookQueryParam)),
		option.Response(200, new(OrderBookResponse)),
	)

	// WebSocket (not documented in OpenAPI)
	app.Use("/ws", ws.UpgradeMiddleware())
	app.Get("/ws", ws.Handler(hub))
}
