package clients

import (
	"context"
	"fmt"

	orderv1 "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients — gRPC клиенты к микросервисам
type Clients struct {
	User    userv1.UserServiceClient
	Product productv1.ProductServiceClient
	Order   orderv1.OrderServiceClient

	conns []*grpc.ClientConn
}

// New создаёт gRPC-клиенты
func New(ctx context.Context, userAddr, productAddr, orderAddr string) (*Clients, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	userConn, err := grpc.NewClient(userAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to user-service: %w", err)
	}

	productConn, err := grpc.NewClient(productAddr, opts...)
	if err != nil {
		userConn.Close()
		return nil, fmt.Errorf("connect to product-service: %w", err)
	}

	orderConn, err := grpc.NewClient(orderAddr, opts...)
	if err != nil {
		userConn.Close()
		productConn.Close()
		return nil, fmt.Errorf("connect to order-service: %w", err)
	}

	return &Clients{
		User:    userv1.NewUserServiceClient(userConn),
		Product: productv1.NewProductServiceClient(productConn),
		Order:   orderv1.NewOrderServiceClient(orderConn),
		conns:   []*grpc.ClientConn{userConn, productConn, orderConn},
	}, nil
}

// Close закрывает соединения
func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}
