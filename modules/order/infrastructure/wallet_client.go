package infrastructure

import (
	"context"

	"booker/modules/order/domain"
	"booker/modules/order/domain/interfaces"
	walletpb "booker/proto/wallet/v1/gen"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type walletClient struct {
	client walletpb.WalletServiceClient
}

func NewWalletClient(conn *grpc.ClientConn) interfaces.WalletClient {
	return &walletClient{client: walletpb.NewWalletServiceClient(conn)}
}

func (c *walletClient) HoldBalance(ctx context.Context, userID, assetID string, amount decimal.Decimal) error {
	_, err := c.client.HoldBalance(ctx, &walletpb.BalanceRequest{
		UserId:  userID,
		AssetId: assetID,
		Amount:  amount.String(),
	})
	if err != nil {
		return mapWalletError(err)
	}
	return nil
}

func (c *walletClient) ReleaseBalance(ctx context.Context, userID, assetID string, amount decimal.Decimal) error {
	_, err := c.client.ReleaseBalance(ctx, &walletpb.BalanceRequest{
		UserId:  userID,
		AssetId: assetID,
		Amount:  amount.String(),
	})
	if err != nil {
		return mapWalletError(err)
	}
	return nil
}

func mapWalletError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return domain.ErrWalletServiceUnavailable.Wrap(err)
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return domain.ErrInsufficientBalance
	case codes.NotFound:
		return domain.ErrWalletNotFound.Wrap(err)
	case codes.Unavailable, codes.DeadlineExceeded:
		return domain.ErrWalletServiceUnavailable.Wrap(err)
	default:
		return domain.ErrWalletServiceUnavailable.Wrap(err)
	}
}
