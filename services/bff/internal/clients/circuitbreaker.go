package clients

import (
	"context"
	"errors"

	orderv1 "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	userv1 "github.com/pppestto/ecommerce-grpc/pb/user/v1"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

func cbSettings(name string) gobreaker.Settings {
	return gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    0,
		Timeout:     30,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: nil,
	}
}

type UserClientWithCB struct {
	inner userv1.UserServiceClient
	cb    *gobreaker.CircuitBreaker
}

func NewUserClientWithCB(inner userv1.UserServiceClient) *UserClientWithCB {
	return &UserClientWithCB{
		inner: inner,
		cb:    gobreaker.NewCircuitBreaker(cbSettings("user-service")),
	}
}

func (c *UserClientWithCB) GetUser(ctx context.Context, req *userv1.GetUserRequest, opts ...grpc.CallOption) (*userv1.User, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.GetUser(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*userv1.User), nil
}

func (c *UserClientWithCB) Login(ctx context.Context, req *userv1.LoginRequest, opts ...grpc.CallOption) (*userv1.LoginResponse, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.Login(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*userv1.LoginResponse), nil
}

func (c *UserClientWithCB) CreateUser(ctx context.Context, req *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.CreateUser(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*userv1.CreateUserResponse), nil
}

func (c *UserClientWithCB) DeleteUser(ctx context.Context, req *userv1.GetUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.DeleteUser(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*emptypb.Empty), nil
}

type ProductClientWithCB struct {
	inner productv1.ProductServiceClient
	cb    *gobreaker.CircuitBreaker
}

func NewProductClientWithCB(inner productv1.ProductServiceClient) *ProductClientWithCB {
	return &ProductClientWithCB{
		inner: inner,
		cb:    gobreaker.NewCircuitBreaker(cbSettings("product-service")),
	}
}

func (c *ProductClientWithCB) GetProduct(ctx context.Context, req *productv1.GetProductRequest, opts ...grpc.CallOption) (*productv1.Product, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.GetProduct(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*productv1.Product), nil
}

func (c *ProductClientWithCB) ListProducts(ctx context.Context, req *productv1.ListProductsRequest, opts ...grpc.CallOption) (*productv1.ListProductsResponse, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.ListProducts(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*productv1.ListProductsResponse), nil
}

func (c *ProductClientWithCB) CreateProduct(ctx context.Context, in *productv1.Product, opts ...grpc.CallOption) (*productv1.Product, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.CreateProduct(ctx, in)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*productv1.Product), nil
}

type OrderClientWithCB struct {
	inner orderv1.OrderServiceClient
	cb    *gobreaker.CircuitBreaker
}

func NewOrderClientWithCB(inner orderv1.OrderServiceClient) *OrderClientWithCB {
	return &OrderClientWithCB{
		inner: inner,
		cb:    gobreaker.NewCircuitBreaker(cbSettings("order-service")),
	}
}

func (c *OrderClientWithCB) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.CreateOrder(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*orderv1.CreateOrderResponse), nil
}

func (c *OrderClientWithCB) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.Order, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.GetOrder(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*orderv1.Order), nil
}

func (c *OrderClientWithCB) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest, opts ...grpc.CallOption) (*orderv1.Order, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.UpdateOrderStatus(ctx, req)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, status.Error(codes.Unavailable, ErrCircuitOpen.Error())
		}
		return nil, err
	}
	return result.(*orderv1.Order), nil
}
