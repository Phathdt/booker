package market

import (
	"booker/pkg/httpserver"
	pb "booker/proto/matching/v1/gen"

	"github.com/gofiber/fiber/v2"
)

type OrderBookLevel struct {
	Price      string `json:"price"`
	Quantity   string `json:"quantity"`
	OrderCount int32  `json:"order_count"`
}

type OrderBookResponse struct {
	PairID string           `json:"pair_id"`
	Bids   []OrderBookLevel `json:"bids"`
	Asks   []OrderBookLevel `json:"asks"`
}

// GetOrderBook returns the current order book depth for a trading pair.
func GetOrderBook(matchingClient pb.MatchingServiceClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pair := c.Params("pair")

		resp, err := matchingClient.GetOrderBook(c.Context(), &pb.GetOrderBookRequest{
			PairId: pair,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "matching engine unavailable")
		}

		bids := make([]OrderBookLevel, len(resp.Bids))
		for i, b := range resp.Bids {
			bids[i] = OrderBookLevel{Price: b.Price, Quantity: b.Quantity, OrderCount: b.OrderCount}
		}

		asks := make([]OrderBookLevel, len(resp.Asks))
		for i, a := range resp.Asks {
			asks[i] = OrderBookLevel{Price: a.Price, Quantity: a.Quantity, OrderCount: a.OrderCount}
		}

		return httpserver.OK(c, OrderBookResponse{
			PairID: resp.PairId,
			Bids:   bids,
			Asks:   asks,
		})
	}
}
