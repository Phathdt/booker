package matching

import (
	"context"
	"errors"

	"booker/modules/matching/domain/interfaces"
	"booker/modules/matching/engine"
	apperrors "booker/pkg/errors"
	pb "booker/proto/matching/v1/gen"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchingServer struct {
	pb.UnimplementedMatchingServiceServer
	matchingSvc interfaces.MatchingService
}

func NewMatchingServer(matchingSvc interfaces.MatchingService) *MatchingServer {
	return &MatchingServer{matchingSvc: matchingSvc}
}

func (s *MatchingServer) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderResponse, error) {
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid price")
	}
	qty, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid quantity")
	}

	side := engine.Side(req.Side)
	if side != engine.SideBuy && side != engine.SideSell {
		return nil, status.Error(codes.InvalidArgument, "side must be buy or sell")
	}

	order := &engine.BookOrder{
		ID:        req.OrderId,
		UserID:    req.UserId,
		PairID:    req.PairId,
		Side:      side,
		Price:     price,
		Quantity:  qty,
		Remaining: qty,
	}

	trades, err := s.matchingSvc.SubmitOrder(ctx, order)
	if err != nil {
		return nil, toGRPCError(err)
	}

	respStatus := "accepted"
	if len(trades) > 0 {
		if order.Remaining.IsZero() {
			respStatus = "filled"
		} else {
			respStatus = "partial"
		}
	}

	tradeResults := make([]*pb.TradeResult, len(trades))
	for i, t := range trades {
		tradeResults[i] = &pb.TradeResult{
			TradeId:     t.ID,
			BuyOrderId:  t.BuyOrderID,
			SellOrderId: t.SellOrderID,
			Price:       t.Price.String(),
			Quantity:    t.Quantity.String(),
		}
	}

	return &pb.SubmitOrderResponse{
		OrderId: req.OrderId,
		Status:  respStatus,
		Trades:  tradeResults,
	}, nil
}

func (s *MatchingServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	if err := s.matchingSvc.CancelOrder(ctx, req.PairId, req.OrderId); err != nil {
		return &pb.CancelOrderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CancelOrderResponse{
		Success: true,
		Message: "order removed from book",
	}, nil
}

func toGRPCError(err error) error {
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, err.Error())
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		code := codes.Internal
		switch appErr.StatusCode() {
		case 400:
			code = codes.InvalidArgument
		case 404:
			code = codes.NotFound
		case 503:
			code = codes.Unavailable
		}
		return status.Error(code, appErr.Message())
	}
	return status.Error(codes.Internal, "internal error")
}
