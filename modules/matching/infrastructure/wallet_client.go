package infrastructure

import (
	"context"

	"booker/modules/matching/domain"
	"booker/modules/matching/domain/interfaces"
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

func (c *walletClient) SettleTrade(ctx context.Context, userID, assetID string, amount decimal.Decimal) error {
	_, err := c.client.SettleTrade(ctx, &walletpb.BalanceRequest{
		UserId:  userID,
		AssetId: assetID,
		Amount:  amount.String(),
	})
	if err != nil {
		return mapWalletError(err)
	}
	return nil
}

func (c *walletClient) Deposit(ctx context.Context, userID, assetID string, amount decimal.Decimal) error {
	_, err := c.client.Deposit(ctx, &walletpb.BalanceRequest{
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
		return domain.ErrSettlementFailed.Wrap(err)
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded:
		return domain.ErrSettlementFailed.Wrap(err)
	default:
		return domain.ErrSettlementFailed.Wrap(err)
	}
}
