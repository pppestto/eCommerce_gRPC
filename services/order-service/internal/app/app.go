package app

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/pkg/errors"
	pb "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	"github.com/pppestto/ecommerce-grpc/pkg/otel"
	"github.com/pppestto/ecommerce-grpc/services/common/logger"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/usecase"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type App struct {
	grpcServer     *grpc.Server
	listener       net.Listener
	relayCancel    context.CancelFunc
	consumerCancel context.CancelFunc
	otelShutdown   func(context.Context) error
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func New() (*App, error) {
	otelShutdown, err := otel.InitTracer("order-service")
	if err != nil {
		return nil, errors.Wrap(err, "init tracer")
	}

	repository, err := postgres.New(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create postgres repository")
	}

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topic := getEnv("KAFKA_ORDER_TOPIC", "order-events")
	producer, err := kafka.NewOrderKafkaProducer(brokers, topic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kafka producer")
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
	go kafka.OutboxRelay(relayCtx, outboxStore, producer, 50, 2*time.Second, logger.L())

	consumer, err := kafka.NewOrderKafkaConsumer(brokers, topic, "order-service-consumer", repository, logger.L())
	if err != nil {
		relayCancel()
		return nil, errors.Wrap(err, "failed to create kafka consumer")
	}
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	orderEventHandler := func(ctx context.Context, ev kafka.OrderEvent, eventType string) error {
		payload, _ := json.Marshal(ev)
		return repository.LogOrderEvent(ctx, ev.ID, eventType, payload)
	}
	go consumer.Consume(consumerCtx, orderEventHandler)

	orderHandler := handler.NewOrderHandler(orderService)

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		return nil, errors.Wrap(err, "failed to listen")
	}

	return &App{
		grpcServer:     grpcServer,
		listener:       listener,
		relayCancel:    relayCancel,
		consumerCancel: consumerCancel,
		otelShutdown:   otelShutdown,
	}, nil
}

func (a *App) Run() error {
	logger.L().Info("order-service starting", "addr", ":50053")
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	_ = a.otelShutdown(context.Background())
	a.consumerCancel()
	a.relayCancel()
	a.grpcServer.GracefulStop()
}
