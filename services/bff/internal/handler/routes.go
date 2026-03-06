package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/auth"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RoutesConfig struct {
	Order      *OrderHandler
	Product    *ProductHandler
	Auth       *AuthHandler
	JWTManager *auth.JWTManager
	Logger     *slog.Logger
}

func Routes(cfg RoutesConfig) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("POST /api/auth/login", cfg.Auth.Login)
	mux.HandleFunc("POST /api/auth/register", cfg.Auth.Register)
	mux.HandleFunc("GET /api/health", health)

	mux.HandleFunc("GET /api/products", cfg.Product.ListProducts)
	mux.HandleFunc("GET /api/products/{id}", cfg.Product.GetProduct)

	authMiddleware := middleware.Auth(cfg.JWTManager)
	mux.Handle("POST /api/orders", authMiddleware(http.HandlerFunc(cfg.Order.CreateOrder)))

	chain := middleware.RequestID(middleware.Logging(cfg.Logger)(mux))
	return chain
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
