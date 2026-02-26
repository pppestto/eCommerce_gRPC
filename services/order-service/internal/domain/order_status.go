package domain

// OrderStatus — статус заказа (доменный тип, не зависит от pb)
type OrderStatus int

const (
	OrderStatusUnspecified OrderStatus = 0
	OrderStatusPending     OrderStatus = 1
	OrderStatusPaid        OrderStatus = 2
	OrderStatusShipped     OrderStatus = 3
	OrderStatusCancelled   OrderStatus = 4
)
