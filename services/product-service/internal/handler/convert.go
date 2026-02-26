package handler

import (
	commonpb "github.com/pppestto/ecommerce-grpc/pb/common"
	pb "github.com/pppestto/ecommerce-grpc/pb/product/v1"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
)

func domainProductToPB(p *domain.Product) *pb.Product {
	if p == nil {
		return nil
	}
	return &pb.Product{
		Id:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price: &commonpb.Money{
			Amount:   p.Price.Amount,
			Currency: p.Price.Currency,
		},
		Stock: p.Stock,
	}
}
