package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/pppestto/ecommerce-grpc/pb/common"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
)

type GetProductRequest struct {
	Id string `json:"id"`
}

type Pagination struct {
	Page int32 `json:"page"`
	Size int32 `json:"size"`
}

type ListProductsRequest struct {
	Pagination *Pagination `json:"pagination,omitempty"`
	Category   string      `json:"category,omitempty"`
}

type ProductHandler struct {
	product productv1.ProductServiceClient
	logger  *slog.Logger
}

func NewProductHandler(product productv1.ProductServiceClient, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{
		product: product,
		logger:  logger,
	}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing id")
		return
	}

	ctx := r.Context()
	resp, err := h.product.GetProduct(ctx, &productv1.GetProductRequest{Id: id})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "product not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	page := int32(1)
	size := int32(20)

	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = int32(p)
		}
	}
	if v := q.Get("size"); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			size = int32(s)
		}
	}

	category := q.Get("category")

	ctx := r.Context()
	resp, err := h.product.ListProducts(ctx, &productv1.ListProductsRequest{
		Pagination: &common.Pagination{Page: page, Size: size},
		Category:   category,
	})

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to list products")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
