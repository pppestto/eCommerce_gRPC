package handler

import (
	"encoding/json"
	"net/http"

	orderv1 "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
)

type OrderHandler struct {
	user    userv1.UserServiceClient
	product productv1.ProductServiceClient
	order   orderv1.OrderServiceClient
}

func NewOrderHandler(user userv1.UserServiceClient, product productv1.ProductServiceClient, order orderv1.OrderServiceClient) *OrderHandler {
	return &OrderHandler{
		user:    user,
		product: product,
		order:   order,
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
