package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/clients"
)

// OrderHandler — HTTP handlers для заказов
type OrderHandler struct {
	clients *clients.Clients
}

// NewOrderHandler создаёт handler
func NewOrderHandler(c *clients.Clients) *OrderHandler {
	return &OrderHandler{clients: c}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
