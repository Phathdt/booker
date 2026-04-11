package users

import (
	"context"

	"booker/modules/users/domain/entities"
	"booker/modules/users/domain/interfaces"
	pb "booker/proto/user/v1/gen"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	userSvc interfaces.UserService
}

func NewUserServer(userSvc interfaces.UserService) *UserServer {
	return &UserServer{userSvc: userSvc}
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.User, error) {
	user, err := s.userSvc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return userToProto(user), nil
}

func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)

	users, total, err := s.userSvc.List(ctx, limit, offset)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, u := range users {
		pbUsers[i] = userToProto(u)
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Total: total,
	}, nil
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
