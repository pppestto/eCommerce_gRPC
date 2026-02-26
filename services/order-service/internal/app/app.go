package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/usecase"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// New создаёт и инициализирует приложение
func New() (*App, error) {
	repository, err := postgres.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres repository: %w", err)
	}

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topic := getEnv("KAFKA_ORDER_TOPIC", "order-events")
	producer, err := kafka.NewOrderKafkaProducer(brokers, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	orderService := usecase.NewOrderService(repository, producer)

	// Инициализация handler (gRPC)
	orderHandler := handler.NewOrderHandler(orderService)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	// Слушаем на порту 50051
	listener, err := net.Listen("tcp", ":50053")
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
	fmt.Println("Order Service starting on :50053")
	return a.grpcServer.Serve(a.listener)
}

// Stop останавливает сервер
func (a *App) Stop() {
	a.grpcServer.GracefulStop()
}
