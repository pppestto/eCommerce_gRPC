package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/pppestto/ecommerce-grpc/pkg/otel"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/auth"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/clients"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/handler"
	"github.com/pppestto/ecommerce-grpc/services/common/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type App struct {
	server       *http.Server
	clients      *clients.Clients
	otelShutdown func(context.Context) error
}

type Config struct {
	Port        int
	UserAddr    string
	ProductAddr string
	OrderAddr   string
	RedisAddr   string
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func New(cfg Config) (*App, error) {
	otelShutdown, err := otel.InitTracer("bff")
	if err != nil {
		return nil, errors.Wrap(err, "init tracer")
	}

	clients, err := clients.New(context.Background(), cfg.UserAddr, cfg.ProductAddr, cfg.OrderAddr, cfg.RedisAddr)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc clients")
	}

	jwtSecret := getEnv("JWT_SECRET", "your-256-bit-secret-key-change-in-production")
	jwtTTL := 24 * time.Hour
	jwtManager := auth.NewJWTManager(jwtSecret, jwtTTL)

	productHandler := handler.NewProductHandler(clients.Product, logger.L())
	orderHandler := handler.NewOrderHandler(clients.User, clients.Product, clients.Order, logger.L())
	authHandler := handler.NewAuthHandler(clients.User, jwtManager, logger.L())
	mux := handler.Routes(handler.RoutesConfig{
		Product:    productHandler,
		Order:      orderHandler,
		Auth:       authHandler,
		JWTManager: jwtManager,
		Logger:     logger.L(),
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: otelhttp.NewHandler(mux, "bff"),
	}

	return &App{
		server:       server,
		clients:      clients,
		otelShutdown: otelShutdown,
	}, nil
}

func (a *App) Run() error {
	logger.L().Info("BFF starting", "addr", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) Stop(ctx context.Context) error {
	_ = a.otelShutdown(ctx)
	a.clients.Close()
	return a.server.Shutdown(ctx)
}
