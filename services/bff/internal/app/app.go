package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/clients"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/handler"
)

type App struct {
	server  *http.Server
	clients *clients.Clients
}

type Config struct {
	Port          int
	UserAddr      string
	ProductAddr   string
	OrderAddr     string
}

// New создаёт приложение
func New(cfg Config) (*App, error) {
	clients, err := clients.New(context.Background(), cfg.UserAddr, cfg.ProductAddr, cfg.OrderAddr)
	if err != nil {
		return nil, fmt.Errorf("create grpc clients: %w", err)
	}

	orderHandler := handler.NewOrderHandler(clients)
	mux := handler.Routes(orderHandler)

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

// Run запускает HTTP сервер
func (a *App) Run() error {
	fmt.Printf("BFF starting on %s\n", a.server.Addr)
	return a.server.ListenAndServe()
}

// Stop останавливает сервер и закрывает соединения
func (a *App) Stop(ctx context.Context) error {
	a.clients.Close()
	return a.server.Shutdown(ctx)
}
