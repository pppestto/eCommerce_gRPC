package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/auth"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/clients"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/handler"
)

type App struct {
	server  *http.Server
	clients *clients.Clients
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
	clients, err := clients.New(context.Background(), cfg.UserAddr, cfg.ProductAddr, cfg.OrderAddr, cfg.RedisAddr)
	if err != nil {
		return nil, fmt.Errorf("create grpc clients: %w", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "your-256-bit-secret-key-change-in-production")
	jwtTTL := 24 * time.Hour
	jwtManager := auth.NewJWTManager(jwtSecret, jwtTTL)

	orderHandler := handler.NewOrderHandler(clients.User, clients.Product, clients.Order)
	authHandler := handler.NewAuthHandler(clients.User, jwtManager)
	mux := handler.Routes(handler.RoutesConfig{
		Order:      orderHandler,
		Auth:       authHandler,
		JWTManager: jwtManager,
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &App{
		server:  server,
		clients: clients,
	}, nil
}

func (a *App) Run() error {
	fmt.Printf("BFF starting on %s\n", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) Stop(ctx context.Context) error {
	a.clients.Close()
	return a.server.Shutdown(ctx)
}
