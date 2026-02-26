package handler

import (
	"github.com/google/uuid"
	commonpb "github.com/pppestto/ecommerce-grpc/pb/common"
	pb "github.com/pppestto/ecommerce-grpc/pb/order/v1"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

// pbItemsToDomain конвертирует []*pb.OrderItem в []domain.OrderItem
func pbItemsToDomain(items []*pb.OrderItem) ([]domain.OrderItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	result := make([]domain.OrderItem, 0, len(items))
	for _, it := range items {
		productID, err := uuid.Parse(it.GetProductId())
		if err != nil {
			return nil, err
		}
		price := pbMoneyToDomain(it.GetPrice())
		item, err := domain.NewOrderItem(productID, int(it.GetQuantity()), price)
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, nil
}

func pbMoneyToDomain(m *commonpb.Money) common.Money {
	if m == nil {
		return common.Money{Amount: 0, Currency: ""}
	}
	return common.Money{Amount: m.GetAmount(), Currency: m.GetCurrency()}
}

// domainOrderToPB конвертирует domain.Order в *pb.Order
func domainOrderToPB(o *domain.Order) *pb.Order {
	if o == nil {
		return nil
	}
	items := make([]*pb.OrderItem, len(o.Items))
	for i := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId: o.Items[i].ProductID.String(),
			Quantity:  int32(o.Items[i].Quantity),
			Price: &commonpb.Money{
				Amount:   o.Items[i].Price.Amount,
				Currency: o.Items[i].Price.Currency,
			},
		}
	}
	return &pb.Order{
		Id:     o.ID.String(),
		UserId: o.UserID.String(),
		Items:  items,
		Total:  &commonpb.Money{Amount: o.Total.Amount, Currency: o.Total.Currency},
		Status: pb.OrderStatus(o.Status),
	}
}
