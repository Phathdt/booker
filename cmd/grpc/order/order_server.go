package order

import (
	"context"
	"errors"
	"time"

	"booker/modules/order/application/dto"
	"booker/modules/order/domain/entities"
	"booker/modules/order/domain/interfaces"
	apperrors "booker/pkg/errors"
	pb "booker/proto/order/v1/gen"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	orderSvc interfaces.OrderService
}

func NewOrderServer(orderSvc interfaces.OrderService) *OrderServer {
	return &OrderServer{orderSvc: orderSvc}
}

func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	order, err := s.orderSvc.GetOrderInternal(ctx, req.OrderId)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

func (s *OrderServer) ListUserOrders(
	ctx context.Context,
	req *pb.ListUserOrdersRequest,
) (*pb.ListUserOrdersResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	orders, err := s.orderSvc.ListOrders(ctx, req.UserId, &dto.ListOrdersDTO{
		PairID: req.PairId,
		Status: req.Status,
		Limit:  limit,
		Offset: req.Offset,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.ListUserOrdersResponse{
		Orders: make([]*pb.OrderResponse, len(orders)),
	}
	for i, o := range orders {
		resp.Orders[i] = toProto(o)
	}
	return resp, nil
}

func (s *OrderServer) UpdateOrderFill(ctx context.Context, req *pb.UpdateOrderFillRequest) (*pb.OrderResponse, error) {
	filledQty, err := decimal.NewFromString(req.FilledQty)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid filled_qty")
	}

	order, err := s.orderSvc.UpdateOrderFill(ctx, req.OrderId, filledQty, req.Status)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

func toProto(o *entities.Order) *pb.OrderResponse {
	return &pb.OrderResponse{
		Id:        o.ID,
		UserId:    o.UserID,
		PairId:    o.PairID,
		Side:      o.Side,
		Type:      o.Type,
		Price:     o.Price.String(),
		Quantity:  o.Quantity.String(),
		FilledQty: o.FilledQty.String(),
		Status:    o.Status,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
	}
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
