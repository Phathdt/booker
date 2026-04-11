package users

import (
	"context"

	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"
	pb "booker/proto/user/v1/gen"

	"google.golang.org/protobuf/types/known/emptypb"
)

// UserServer implements the inter-service gRPC UserService.
type UserServer struct {
	pb.UnimplementedUserServiceServer
	userSvc  interfaces.UserService
	tokenSvc interfaces.TokenService
}

func NewUserServer(userSvc interfaces.UserService, tokenSvc interfaces.TokenService) *UserServer {
	return &UserServer{userSvc: userSvc, tokenSvc: tokenSvc}
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.User, error) {
	user, err := s.userSvc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return userToProto(user), nil
}

func (s *UserServer) ValidateAccessToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.TokenClaims, error) {
	claims, err := s.tokenSvc.ValidateAccessToken(ctx, req.Token)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.TokenClaims{
		UserId: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Jti:    claims.JTI,
	}, nil
}

func (s *UserServer) RevokeAllUserTokens(ctx context.Context, req *pb.RevokeAllRequest) (*emptypb.Empty, error) {
	if err := s.tokenSvc.RevokeAllUserTokens(ctx, req.UserId); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func userToProto(u *entities.User) *pb.User {
	return &pb.User{
		Id:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
