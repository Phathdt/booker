package market

import (
	"booker/pkg/httpserver"
	pb "booker/proto/matching/v1/gen"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderBookLevel struct {
	Price      string `json:"price"       required:"true" example:"50000.00"`
	Quantity   string `json:"quantity"    required:"true" example:"1.5"`
	OrderCount int32  `json:"orderCount" required:"true" example:"3"`
}

type OrderBookResponse struct {
	PairID string           `json:"pairId" required:"true" example:"BTC_USDT"`
	Bids   []OrderBookLevel `json:"bids"    required:"true"`
	Asks   []OrderBookLevel `json:"asks"    required:"true"`
}

// GetOrderBook godoc
func GetOrderBook(matchingClient pb.MatchingServiceClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pair := c.Params("pair")

		depth := int32(c.QueryInt("depth", 20))
		if depth > 100 {
			depth = 100
		}

		resp, err := matchingClient.GetOrderBook(c.UserContext(), &pb.GetOrderBookRequest{
			PairId: pair,
			Depth:  depth,
		})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				return fiber.NewError(fiber.StatusNotFound, "trading pair not found")
			}
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
