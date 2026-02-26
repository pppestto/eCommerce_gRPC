package handler

import (
	"context"
	"errors"

	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	pb "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one order item is required")
	}

	domainItems, err := pbItemsToDomain(req.Items)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order item: "+err.Error())
	}

	order, err := h.service.CreateOrder(ctx, req.UserId, domainItems)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.CreateOrderResponse{
		Order: domainOrderToPB(order),
	}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	order, err := h.service.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return domainOrderToPB(order), nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.Order, error) {
	if req == nil || req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.NewStatus == pb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "new_status is required")
	}

	order, err := h.service.UpdateOrderStatus(ctx, req.OrderId, domain.OrderStatus(req.NewStatus))
	if err != nil {
		return nil, mapServiceError(err)
	}

	return domainOrderToPB(order), nil
}

// mapServiceError маппит доменные/use case ошибки в gRPC status
func mapServiceError(err error) error {
	if errors.Is(err, commonerrors.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, commonerrors.ErrInvalidArgument) || errors.Is(err, commonerrors.ErrAlreadyExists) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
