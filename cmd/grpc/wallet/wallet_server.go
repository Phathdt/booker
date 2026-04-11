package wallet

import (
	"context"
	"errors"
	"time"

	"booker/modules/wallet/domain/entities"
	"booker/modules/wallet/domain/interfaces"
	apperrors "booker/pkg/errors"
	pb "booker/proto/wallet/v1/gen"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WalletServer struct {
	pb.UnimplementedWalletServiceServer
	walletSvc interfaces.WalletService
}

func NewWalletServer(walletSvc interfaces.WalletService) *WalletServer {
	return &WalletServer{walletSvc: walletSvc}
}

func (s *WalletServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.WalletResponse, error) {
	w, err := s.walletSvc.GetBalance(ctx, req.UserId, req.AssetId)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(w), nil
}

func (s *WalletServer) HoldBalance(ctx context.Context, req *pb.BalanceRequest) (*pb.WalletResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	w, err := s.walletSvc.HoldBalance(ctx, req.UserId, req.AssetId, amount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(w), nil
}

func (s *WalletServer) ReleaseBalance(ctx context.Context, req *pb.BalanceRequest) (*pb.WalletResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	w, err := s.walletSvc.ReleaseBalance(ctx, req.UserId, req.AssetId, amount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(w), nil
}

func (s *WalletServer) SettleTrade(ctx context.Context, req *pb.BalanceRequest) (*pb.WalletResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	w, err := s.walletSvc.SettleTrade(ctx, req.UserId, req.AssetId, amount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(w), nil
}

func (s *WalletServer) Deposit(ctx context.Context, req *pb.BalanceRequest) (*pb.WalletResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	w, err := s.walletSvc.Deposit(ctx, req.UserId, req.AssetId, amount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(w), nil
}

func toProto(w *entities.Wallet) *pb.WalletResponse {
	return &pb.WalletResponse{
		Id:        w.ID,
		UserId:    w.UserID,
		AssetId:   w.AssetID,
		Available: w.Available.String(),
		Locked:    w.Locked.String(),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}

func toGRPCError(err error) error {
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		code := codes.Internal
		switch appErr.StatusCode() {
		case 400:
			code = codes.InvalidArgument
		case 404:
			code = codes.NotFound
		}
		return status.Error(code, appErr.Message())
	}
	return status.Error(codes.Internal, "internal error")
}
