package clients

import (
	"context"
	"encoding/json"
	"time"

	productv1 "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
)

type productCache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

func productItemKey(id string) string {
	return "bff:product:item:" + id
}

type CachedProductClient struct {
	inner   productv1.ProductServiceClient
	cache   productCache
	ttl     time.Duration
	sfGroup singleflight.Group
}

func NewCachedProductClient(inner productv1.ProductServiceClient, c productCache, ttl time.Duration) *CachedProductClient {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &CachedProductClient{
		inner: inner,
		cache: c,
		ttl:   ttl,
	}
}

func (c *CachedProductClient) GetProduct(ctx context.Context, req *productv1.GetProductRequest, opts ...grpc.CallOption) (*productv1.Product, error) {
	if req.Id == "" {
		return c.inner.GetProduct(ctx, req, opts...)
	}

	key := productItemKey(req.Id)

	data, found, err := c.cache.Get(ctx, key)
	if err != nil {
		return c.inner.GetProduct(ctx, req, opts...)
	}
	if found {
		var p productv1.Product
		if err := json.Unmarshal(data, &p); err != nil {
			return c.inner.GetProduct(ctx, req, opts...)
		}
		return &p, nil
	}

	result, err, _ := c.sfGroup.Do(key, func() (interface{}, error) {
		data, found, _ := c.cache.Get(ctx, key)
		if found {
			var p productv1.Product
			if err := json.Unmarshal(data, &p); err == nil {
				return &p, nil
			}
		}

		product, err := c.inner.GetProduct(ctx, req, opts...)
		if err != nil {
			return nil, err
		}

		if data, err := json.Marshal(product); err == nil {
			_ = c.cache.Set(ctx, key, data, c.ttl)
		}

		return product, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(*productv1.Product), nil
}

func (c *CachedProductClient) ListProducts(ctx context.Context, req *productv1.ListProductsRequest, opts ...grpc.CallOption) (*productv1.ListProductsResponse, error) {
	return c.inner.ListProducts(ctx, req, opts...)
}

func (c *CachedProductClient) CreateProduct(ctx context.Context, in *productv1.Product, opts ...grpc.CallOption) (*productv1.Product, error) {
	product, err := c.inner.CreateProduct(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return product, nil
}
