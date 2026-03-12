package domain

import (
	"time"
)

type OrderStatus string



type ShippingStatus string

const (
	ShippingStatusPending       ShippingStatus = "pending"
	ShippingStatusAwaitingPickup ShippingStatus = "awaiting_pickup"
	ShippingStatusShipped       ShippingStatus = "shipped"
	ShippingStatusInTransit     ShippingStatus = "in_transit"
	ShippingStatusDelivered     ShippingStatus = "delivered"
	ShippingStatusCancelled     ShippingStatus = "cancelled"
)


type OrderFilter struct {
	MerchantID     int64          `json:"merchant_id"`
	ShopID         string         `json:"shop_id"`
	Status         OrderStatus    `json:"status"`
	ShippingStatus ShippingStatus `json:"shipping_status"`
	FromDate       *time.Time     `json:"from_date"`
	ToDate         *time.Time     `json:"to_date"`
	Limit          int            `json:"limit"`
	Offset         int            `json:"offset"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required,oneof=new processing packed shipped cancelled"`
}

type CancelOrderRequest struct {
	OrderSN string `json:"order_sn" validate:"required"`
	Reason  string `json:"reason"`
}
