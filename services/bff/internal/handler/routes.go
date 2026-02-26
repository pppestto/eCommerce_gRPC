package handler

import (
	"encoding/json"
	"net/http"
)

// Routes возвращает http.Handler с маршрутами
func Routes(order *OrderHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/orders", order.CreateOrder)
	mux.HandleFunc("GET /api/health", health)

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
