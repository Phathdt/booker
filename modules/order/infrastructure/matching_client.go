package infrastructure

import (
	"context"

	"booker/modules/order/domain/entities"
	"booker/modules/order/domain/interfaces"
	matchingpb "booker/proto/matching/v1/gen"

	"google.golang.org/grpc"
)

type matchingClient struct {
	client matchingpb.MatchingServiceClient
}

func NewMatchingClient(conn *grpc.ClientConn) interfaces.MatchingClient {
	return &matchingClient{client: matchingpb.NewMatchingServiceClient(conn)}
}

func (c *matchingClient) SubmitOrder(ctx context.Context, order *entities.Order) error {
	_, err := c.client.SubmitOrder(ctx, &matchingpb.SubmitOrderRequest{
		OrderId:  order.ID,
		UserId:   order.UserID,
		PairId:   order.PairID,
		Side:     order.Side,
		Price:    order.Price.String(),
		Quantity: order.Quantity.String(),
	})
	return err
}

func (c *matchingClient) CancelOrder(ctx context.Context, pairID, orderID string) error {
	resp, err := c.client.CancelOrder(ctx, &matchingpb.CancelOrderRequest{
		OrderId: orderID,
		PairId:  pairID,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return nil // order not in book is ok (might be already matched)
	}
	return nil
}
