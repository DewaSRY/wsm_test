package domain

import (
	"time"
)

type OrderStatus string

// const (
// 	OrderStatusNew        OrderStatus = "new"
// 	OrderStatusProcessing OrderStatus = "processing"
// 	OrderStatusPacked     OrderStatus = "packed"
// 	OrderStatusShipped    OrderStatus = "shipped"
// 	OrderStatusDelivered  OrderStatus = "delivered"
// 	OrderStatusCancelled  OrderStatus = "cancelled"
// )

type ShippingStatus string

const (
	ShippingStatusPending       ShippingStatus = "pending"
	ShippingStatusAwaitingPickup ShippingStatus = "awaiting_pickup"
	ShippingStatusShipped       ShippingStatus = "shipped"
	ShippingStatusInTransit     ShippingStatus = "in_transit"
	ShippingStatusDelivered     ShippingStatus = "delivered"
	ShippingStatusCancelled     ShippingStatus = "cancelled"
)

// type Order struct {
// 	bun.BaseModel `bun:"table:orders,alias:o"`

// 	ID              int64          `bun:"id,pk,autoincrement" json:"id"`
// 	MerchantID      int64          `bun:"merchant_id,notnull" json:"merchant_id"`
// 	OrderSN         string         `bun:"order_sn,notnull,unique" json:"order_sn"`
// 	ShopID          string         `bun:"shop_id,notnull" json:"shop_id"`
// 	Status          OrderStatus    `bun:"status,notnull,default:'new'" json:"status"`
// 	ShippingStatus  ShippingStatus `bun:"shipping_status,notnull,default:'pending'" json:"shipping_status"`
// 	TotalAmount     float64        `bun:"total_amount,notnull" json:"total_amount"`
// 	Currency        string         `bun:"currency,notnull,default:'IDR'" json:"currency"`
// 	CustomerName    string         `bun:"customer_name" json:"customer_name"`
// 	CustomerPhone   string         `bun:"customer_phone" json:"customer_phone"`
// 	CustomerAddress string         `bun:"customer_address" json:"customer_address"`
// 	Notes           string         `bun:"notes" json:"notes"`
// 	MarketplaceData string         `bun:"marketplace_data,type:jsonb" json:"marketplace_data,omitempty"`
// 	SyncedAt        *time.Time     `bun:"synced_at" json:"synced_at,omitempty"`
// 	CreatedAt       time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
// 	UpdatedAt       time.Time      `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

// 	Merchant   *Merchant    `bun:"rel:belongs-to,join:merchant_id=id" json:"merchant,omitempty"`
// 	Items      []*OrderItem `bun:"rel:has-many,join:id=order_id" json:"items,omitempty"`
// 	Shipment   *Shipment    `bun:"rel:has-one,join:id=order_id" json:"shipment,omitempty"`
// }

// type OrderItem struct {
// 	bun.BaseModel `bun:"table:order_items,alias:oi"`

// 	ID        int64   `bun:"id,pk,autoincrement" json:"id"`
// 	OrderID   int64   `bun:"order_id,notnull" json:"order_id"`
// 	SKU       string  `bun:"sku,notnull" json:"sku"`
// 	Name      string  `bun:"name" json:"name"`
// 	Quantity  int     `bun:"quantity,notnull" json:"quantity"`
// 	Price     float64 `bun:"price,notnull" json:"price"`
// 	Subtotal  float64 `bun:"subtotal,notnull" json:"subtotal"`
// 	Weight    float64 `bun:"weight" json:"weight"`
// 	ImageURL  string  `bun:"image_url" json:"image_url"`

// 	Order *Order `bun:"rel:belongs-to,join:order_id=id" json:"order,omitempty"`
// }

type OrderListResponse struct {
	Message string         `json:"message"`
	Data    []MarketplaceOrder `json:"data"`
}

type MarketplaceOrder struct {
	OrderSN        string              `json:"order_sn"`
	ShopID         string              `json:"shop_id"`
	Status         string              `json:"status"`
	ShippingStatus string              `json:"shipping_status"`
	Items          []MarketplaceOrderItem `json:"items"`
	TotalAmount    float64             `json:"total_amount"`
}

type MarketplaceOrderItem struct {
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

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
