package handler

import (
	"encoding/json"
	"net/http"

	commonpb "github.com/pppestto/ecommerce-grpc/pb/common"
	orderv1 "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateOrderRequest — REST модель для создания заказа
type CreateOrderRequest struct {
	UserID string              `json:"user_id"`
	Items  []CreateOrderItem   `json:"items"`
}

type CreateOrderItem struct {
	ProductID string    `json:"product_id"`
	Quantity  int32     `json:"quantity"`
	Price     MoneyJSON `json:"price"`
}

type MoneyJSON struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if len(req.Items) == 0 {
		writeJSONError(w, http.StatusBadRequest, "at least one item is required")
		return
	}

	ctx := r.Context()

	// 1. Проверяем существование пользователя
	_, err := h.clients.User.GetUser(ctx, &userv1.GetUserRequest{Id: req.UserID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			writeJSONError(w, http.StatusBadRequest, "user not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to validate user")
		return
	}

	// 2. Проверяем существование каждого продукта
	for _, item := range req.Items {
		if item.ProductID == "" || item.Quantity <= 0 || item.Price.Amount < 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid item")
			return
		}
		_, err := h.clients.Product.GetProduct(ctx, &productv1.GetProductRequest{Id: item.ProductID})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				writeJSONError(w, http.StatusBadRequest, "product not found: "+item.ProductID)
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "failed to validate product")
			return
		}
	}

	// 3. Конвертируем в gRPC запрос
	items := make([]*orderv1.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = &orderv1.OrderItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
			Price:     &commonpb.Money{Amount: it.Price.Amount, Currency: it.Price.Currency},
		}
	}

	grpcReq := &orderv1.CreateOrderRequest{
		UserId: req.UserID,
		Items:  items,
	}

	resp, err := h.clients.Order.CreateOrder(ctx, grpcReq)
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			writeJSONError(w, http.StatusBadRequest, status.Convert(err).Message())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp.Order)
}
