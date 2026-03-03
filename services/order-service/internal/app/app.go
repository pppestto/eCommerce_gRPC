package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/usecase"
)

type App struct {
	grpcServer     *grpc.Server
	listener       net.Listener
	relayCancel    context.CancelFunc
	consumerCancel context.CancelFunc
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

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

	orderService := usecase.NewOrderService(repository)

	outboxStore := kafka.NewOutboxStore(
		func(ctx context.Context, limit int) ([]kafka.OutboxRow, error) {
			rows, err := repository.GetUnpublishedOutbox(ctx, limit)
			if err != nil {
				return nil, err
			}
			result := make([]kafka.OutboxRow, len(rows))
			for i, r := range rows {
				result[i] = kafka.OutboxRow{
					ID:        r.ID,
					EventType: r.EventType,
					Payload:   r.Payload,
				}
			}
			return result, nil
		},
		repository.MarkOutboxPublished,
	)

	relayCtx, relayCancel := context.WithCancel(context.Background())
	go kafka.OutboxRelay(relayCtx, outboxStore, producer, 50, 2*time.Second)

	consumer, err := kafka.NewOrderKafkaConsumer(brokers, topic, "order-service-consumer", repository)
	if err != nil {
		relayCancel()
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	orderEventHandler := func(ctx context.Context, ev kafka.OrderEvent, eventType string) error {
		payload, _ := json.Marshal(ev)
		return repository.LogOrderEvent(ctx, ev.ID, eventType, payload)
	}
	go consumer.Consume(consumerCtx, orderEventHandler)

	orderHandler := handler.NewOrderHandler(orderService)

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &App{
		grpcServer:     grpcServer,
		listener:       listener,
		relayCancel:    relayCancel,
		consumerCancel: consumerCancel,
	}, nil
}

func (a *App) Run() error {
	fmt.Println("Order Service starting on :50053")
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	a.consumerCancel()
	a.relayCancel()
	a.grpcServer.GracefulStop()
}
