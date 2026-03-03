package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/auth"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/middleware"
)

type RoutesConfig struct {
	Order      *OrderHandler
	Auth       *AuthHandler
	JWTManager *auth.JWTManager
}

func Routes(cfg RoutesConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/login", cfg.Auth.Login)
	mux.HandleFunc("POST /api/auth/register", cfg.Auth.Register)
	mux.HandleFunc("GET /api/health", health)

	authMiddleware := middleware.Auth(cfg.JWTManager)
	mux.Handle("POST /api/orders", authMiddleware(http.HandlerFunc(cfg.Order.CreateOrder)))

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
