package app

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	// "github.com/pppestto/ecommerce-grpc/services/product-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/usecase"
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
	// eventBus, err := kafka.New()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create kafka event bus: %w", err)
	// }

	// Инициализация usecase слоя
	productService := usecase.NewProductService(repository)

	// Инициализация handler (gRPC)
	productHandler := handler.NewProductHandler(productService)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, productHandler)
	reflection.Register(grpcServer)

	// Слушаем на порту 50052 (50051 — user-service)
	listener, err := net.Listen("tcp", ":50052")
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
	fmt.Println("Product Service starting on :50052")
	return a.grpcServer.Serve(a.listener)
}

// Stop останавливает сервер
func (a *App) Stop() {
	a.grpcServer.GracefulStop()
}
