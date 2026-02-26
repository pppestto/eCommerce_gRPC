package app

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/usecase"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// New создаёт и инициализирует приложение
func New() (*App, error) {
	// Инициализация хранилища
	repository, err := postgres.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres repository: %w", err)
	}

	// ИнициализацияEventBus (Kafka)
	eventBus, err := kafka.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka event bus: %w", err)
	}

	// Инициализация usecase слоя
	userService := usecase.NewUserService(repository, eventBus)

	// Инициализация handler (gRPC)
	userHandler := handler.NewUserHandler(userService)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	// Слушаем на порту 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &App{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

// Run запускает gRPC сервер
func (a *App) Run() error {
	fmt.Println("User Service starting on :50051")
	return a.grpcServer.Serve(a.listener)
}

// Stop останавливает сервер
func (a *App) Stop() {
	a.grpcServer.GracefulStop()
}
