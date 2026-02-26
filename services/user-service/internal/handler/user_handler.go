package handler

import (
	"context"

	pb "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service UserService
}

// NewUserHandler создаёт новый handler
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// CreateUser создаёт нового пользователя
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	user, err := h.service.CreateUser(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:    user.ID,
			Email: user.Email,
		},
	}, nil
}

// GetUser получает пользователя по ID
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, err := h.service.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.User{
		Id:    user.ID,
		Email: user.Email,
	}, nil
}

// DeleteUser удаляет пользователя
func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.GetUserRequest) (*emptypb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.service.DeleteUser(ctx, req.Id); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &emptypb.Empty{}, nil
}
