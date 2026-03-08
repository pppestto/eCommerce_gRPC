package handler

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/pppestto/ecommerce-grpc/pb/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentHandler implements payment.v1.PaymentService (mock: всегда успешная оплата).
type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

func (h *PaymentHandler) Pay(ctx context.Context, req *pb.PayRequest) (*pb.PayResponse, error) {
	if req == nil || req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount == nil || req.Amount.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount is required and must be positive")
	}

	// Mock: всегда успех, возвращаем id транзакции
	txID := "tx-" + uuid.New().String()
	return &pb.PayResponse{
		Success:       true,
		TransactionId: txID,
	}, nil
}
