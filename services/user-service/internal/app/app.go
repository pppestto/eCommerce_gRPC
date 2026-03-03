package app

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/auth"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/usecase"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func New() (*App, error) {
	repository, err := postgres.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres repository: %w", err)
	}

	eventBus, err := kafka.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka event bus: %w", err)
	}

	passwordHasher := auth.NewBcryptHasher()
	userService := usecase.NewUserService(repository, eventBus, passwordHasher)

	userHandler := handler.NewUserHandler(userService)

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &App{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

func (a *App) Run() error {
	fmt.Println("User Service starting on :50051")
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	a.grpcServer.GracefulStop()
}
