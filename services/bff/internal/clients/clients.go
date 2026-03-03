package clients

import (
	"context"
	"fmt"
	"time"

	orderv1 "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/pppestto/ecommerce-grpc/services/bff/internal/cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	User    userv1.UserServiceClient
	Product productv1.ProductServiceClient
	Order   orderv1.OrderServiceClient

	conns      []*grpc.ClientConn
	redisCache *cache.RedisCache
}

func New(ctx context.Context, userAddr, productAddr, orderAddr, redisAddr string) (*Clients, error) {
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

	userRaw := userv1.NewUserServiceClient(userConn)
	productRaw := productv1.NewProductServiceClient(productConn)
	orderRaw := orderv1.NewOrderServiceClient(orderConn)

	var productClient productv1.ProductServiceClient = NewProductClientWithCB(productRaw)
	if redisAddr != "" {
		redisCache, err := cache.NewRedisCache(redisAddr)
		if err == nil {
			productClient = NewCachedProductClient(productClient, redisCache, 15*time.Minute)
			return &Clients{
				User:       NewUserClientWithCB(userRaw),
				Product:    productClient,
				Order:      NewOrderClientWithCB(orderRaw),
				conns:      []*grpc.ClientConn{userConn, productConn, orderConn},
				redisCache: redisCache,
			}, nil
		}
	}

	return &Clients{
		User:    NewUserClientWithCB(userRaw),
		Product: productClient,
		Order:   NewOrderClientWithCB(orderRaw),
		conns:   []*grpc.ClientConn{userConn, productConn, orderConn},
	}, nil
}

func (c *Clients) Close() {
	if c.redisCache != nil {
		_ = c.redisCache.Close()
	}
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}
