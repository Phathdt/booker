package users

import (
	"context"
	"fmt"

	"booker/modules/users/application/dto"
	"booker/modules/users/application/usecases"
	"booker/modules/users/domain/interfaces"
	"booker/pkg/interceptors"
	pb "booker/proto/user/v1/gen"

	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	registerUC     *usecases.RegisterUseCase
	loginUC        *usecases.LoginUseCase
	refreshTokenUC *usecases.RefreshTokenUseCase
	logoutUC       *usecases.LogoutUseCase
	tokenSvc       interfaces.TokenService
	userSvc        interfaces.UserService
}

func NewAuthServer(
	registerUC *usecases.RegisterUseCase,
	loginUC *usecases.LoginUseCase,
	refreshTokenUC *usecases.RefreshTokenUseCase,
	logoutUC *usecases.LogoutUseCase,
	tokenSvc interfaces.TokenService,
	userSvc interfaces.UserService,
) *AuthServer {
	return &AuthServer{
		registerUC:     registerUC,
		loginUC:        loginUC,
		refreshTokenUC: refreshTokenUC,
		logoutUC:       logoutUC,
		tokenSvc:       tokenSvc,
		userSvc:        userSvc,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	result, err := s.registerUC.Execute(ctx, dto.RegisterDTO{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AuthResponse{
		User:         userToProto(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	result, err := s.loginUC.Execute(ctx, dto.LoginDTO{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AuthResponse{
		User:         userToProto(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenPairResponse, error) {
	result, err := s.refreshTokenUC.Execute(ctx, req.RefreshToken)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.TokenPairResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	userID := req.UserId
	if userID == "" {
		userID = interceptors.UserIDFromCtx(ctx)
	}
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if err := s.logoutUC.Execute(ctx, userID); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *AuthServer) GetMe(ctx context.Context, _ *pb.GetMeRequest) (*pb.User, error) {
	userID := interceptors.UserIDFromCtx(ctx)
	if userID == "" {
		return nil, fmt.Errorf("x-user-id header is required")
	}

	user, err := s.userSvc.GetByID(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return userToProto(user), nil
}

func (s *AuthServer) ValidateAccessToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.TokenClaims, error) {
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

func (s *AuthServer) RevokeAllUserTokens(ctx context.Context, req *pb.RevokeAllRequest) (*emptypb.Empty, error) {
	if err := s.tokenSvc.RevokeAllUserTokens(ctx, req.UserId); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
