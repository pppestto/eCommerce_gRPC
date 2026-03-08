package payment

import (
	"context"

	commonpb "github.com/pppestto/ecommerce-grpc/pb/common"
	paymentv1 "github.com/pppestto/ecommerce-grpc/pb/payment/v1"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/usecase"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcClient реализует usecase.PaymentClient через gRPC PaymentService.
type GrpcClient struct {
	client paymentv1.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcClient(ctx context.Context, addr string) (*GrpcClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		return nil, err
	}
	return &GrpcClient{
		client: paymentv1.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *GrpcClient) Pay(ctx context.Context, orderID string, amount int64, currency string) (bool, error) {
	resp, err := c.client.Pay(ctx, &paymentv1.PayRequest{
		OrderId: orderID,
		Amount:  &commonpb.Money{Amount: amount, Currency: currency},
	})
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func (c *GrpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

var _ usecase.PaymentClient = (*GrpcClient)(nil)
