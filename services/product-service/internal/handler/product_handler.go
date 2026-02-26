package handler

import (
	"context"
	"errors"

	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
	pb "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	product, err := h.service.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return domainProductToPB(product), nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	page, size := int32(1), int32(20)
	if req != nil && req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Size > 0 {
			size = req.Pagination.Size
		}
	}

	category := ""
	if req != nil {
		category = req.Category
	}

	products, total, err := h.service.ListProducts(ctx, int(page), int(size), category)
	if err != nil {
		return nil, mapServiceError(err)
	}

	pbProducts := make([]*pb.Product, len(products))
	for i, p := range products {
		pbProducts[i] = domainProductToPB(p)
	}

	return &pb.ListProductsResponse{
		Products:   pbProducts,
		TotalCount: int32(total),
	}, nil
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	if req == nil || req.Name == "" || req.Price == nil {
		return nil, status.Error(codes.InvalidArgument, "product name and price are required")
	}

	price := common.Money{
		Amount:   req.Price.GetAmount(),
		Currency: req.Price.GetCurrency(),
	}

	product := domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       price,
		Stock:       req.Stock,
	}

	resProduct, err := h.service.CreateProduct(ctx, product)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return domainProductToPB(resProduct), nil
}

func mapServiceError(err error) error {
	if errors.Is(err, commonerrors.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, commonerrors.ErrInvalidArgument) || errors.Is(err, commonerrors.ErrAlreadyExists) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
