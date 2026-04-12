package infrastructure

import (
	"context"

	"booker/modules/matching/domain"
	"booker/modules/matching/domain/interfaces"
	orderpb "booker/proto/order/v1/gen"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type orderClient struct {
	client orderpb.OrderServiceClient
}

func NewOrderClient(conn *grpc.ClientConn) interfaces.OrderClient {
	return &orderClient{client: orderpb.NewOrderServiceClient(conn)}
}

func (c *orderClient) UpdateOrderFill(
	ctx context.Context,
	orderID string,
	filledQty decimal.Decimal,
	orderStatus string,
) error {
	_, err := c.client.UpdateOrderFill(ctx, &orderpb.UpdateOrderFillRequest{
		OrderId:   orderID,
		FilledQty: filledQty.String(),
		Status:    orderStatus,
	})
	if err != nil {
		return mapOrderError(err)
	}
	return nil
}

func mapOrderError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return domain.ErrOrderUpdateFailed.Wrap(err)
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded:
		return domain.ErrOrderUpdateFailed.Wrap(err)
	default:
		return domain.ErrOrderUpdateFailed.Wrap(err)
	}
}
