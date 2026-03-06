package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	commonpb "github.com/pppestto/ecommerce-grpc/pb/common"
	orderv1 "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateOrderRequest struct {
	Items []CreateOrderItem `json:"items"`
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

type OrderHandler struct {
	user    userv1.UserServiceClient
	product productv1.ProductServiceClient
	order   orderv1.OrderServiceClient
	logger  *slog.Logger
}

func NewOrderHandler(user userv1.UserServiceClient, product productv1.ProductServiceClient, order orderv1.OrderServiceClient, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		user:    user,
		product: product,
		order:   order,
		logger:  logger,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Items) == 0 {
		writeJSONError(w, http.StatusBadRequest, "at least one item is required")
		return
	}

	ctx := r.Context()

	_, err := h.user.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			writeJSONError(w, http.StatusBadRequest, "user not found")
			return
		}
		h.logger.Error("failed to validate user", "error", err, "user_id", userID)
		writeJSONError(w, http.StatusInternalServerError, "failed to validate user")
		return
	}

	for _, item := range req.Items {
		if item.ProductID == "" || item.Quantity <= 0 || item.Price.Amount < 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid item")
			return
		}
		_, err := h.product.GetProduct(ctx, &productv1.GetProductRequest{Id: item.ProductID})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				writeJSONError(w, http.StatusBadRequest, "product not found: "+item.ProductID)
				return
			}
			h.logger.Error("failed to validate product", "error", err, "product_id", item.ProductID)
			writeJSONError(w, http.StatusInternalServerError, "failed to validate product")
			return
		}
	}

	items := make([]*orderv1.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = &orderv1.OrderItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
			Price:     &commonpb.Money{Amount: it.Price.Amount, Currency: it.Price.Currency},
		}
	}

	grpcReq := &orderv1.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	}

	resp, err := h.order.CreateOrder(ctx, grpcReq)
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			writeJSONError(w, http.StatusBadRequest, status.Convert(err).Message())
			return
		}
		h.logger.Error("failed to create order", "error", err, "user_id", userID)
		writeJSONError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp.Order)
}
