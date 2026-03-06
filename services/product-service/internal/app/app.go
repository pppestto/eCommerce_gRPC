package app

import (
	"context"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	"github.com/pkg/errors"
	"github.com/pppestto/ecommerce-grpc/pkg/otel"
	"github.com/pppestto/ecommerce-grpc/services/common/logger"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/usecase"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

type App struct {
	grpcServer   *grpc.Server
	listener     net.Listener
	otelShutdown func(context.Context) error
}

// New создаёт и инициализирует приложение
func New() (*App, error) {
	otelShutdown, err := otel.InitTracer("product-service")
	if err != nil {
		return nil, errors.Wrap(err, "init tracer")
	}

	repository, err := postgres.New(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create postgres repository")
	}

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topic := getEnv("KAFKA_PRODUCT_TOPIC", "product-events")
	producer, err := kafka.NewProductKafkaProducer(brokers, topic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kafka event bus")
	}

	productService := usecase.NewProductService(repository, producer)

	productHandler := handler.NewProductHandler(productService)

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterProductServiceServer(grpcServer, productHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		return nil, errors.Wrap(err, "failed to listen")
	}

	return &App{
		grpcServer:   grpcServer,
		listener:     listener,
		otelShutdown: otelShutdown,
	}, nil
}

func (a *App) Run() error {
	logger.L().Info("product-service starting", "addr", ":50052")
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	_ = a.otelShutdown(context.Background())
	a.grpcServer.GracefulStop()
}
