package app

import (
	"context"
	"net"
	"os"

	"github.com/pkg/errors"
	pb "github.com/pppestto/ecommerce-grpc/pb/payment/v1"
	"github.com/pppestto/ecommerce-grpc/pkg/otel"
	"github.com/pppestto/ecommerce-grpc/services/common/logger"
	"github.com/pppestto/ecommerce-grpc/services/payment-service/internal/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

func New() (*App, error) {
	otelShutdown, err := otel.InitTracer("payment-service")
	if err != nil {
		return nil, errors.Wrap(err, "init tracer")
	}

	paymentHandler := handler.NewPaymentHandler()

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterPaymentServiceServer(grpcServer, paymentHandler)
	reflection.Register(grpcServer)

	addr := ":" + getEnv("PAYMENT_GRPC_PORT", "50054")
	listener, err := net.Listen("tcp", addr)
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
	logger.L().Info("payment-service starting", "addr", a.listener.Addr().String())
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	_ = a.otelShutdown(context.Background())
	a.grpcServer.GracefulStop()
}
