package app

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/pkg/errors"
	"github.com/pppestto/ecommerce-grpc/pkg/otel"
	"github.com/pppestto/ecommerce-grpc/services/common/logger"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/auth"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/event/kafka"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/adapters/storage/postgres"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/usecase"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type App struct {
	grpcServer   *grpc.Server
	listener     net.Listener
	otelShutdown func(context.Context) error
}

func New() (*App, error) {
	otelShutdown, err := otel.InitTracer("user-service")
	if err != nil {
		return nil, errors.Wrap(err, "init tracer")
	}

	repository, err := postgres.New(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create postgres repository")
	}

	eventBus, err := kafka.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kafka event bus")
	}

	passwordHasher := auth.NewBcryptHasher()
	userService := usecase.NewUserService(repository, eventBus, passwordHasher)

	userHandler := handler.NewUserHandler(userService)

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50051")
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
	logger.L().Info("user-service starting", "addr", ":50051")
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	_ = a.otelShutdown(context.Background())
	a.grpcServer.GracefulStop()
}
