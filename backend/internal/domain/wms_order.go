package domain

import (
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type WMSStatus string

const (
	WMSStatusReadyToPick WMSStatus = "READY_TO_PICK"
	WMSStatusPicking     WMSStatus = "PICKING"
	WMSStatusPacked      WMSStatus = "PACKED"
	WMSStatusShipped     WMSStatus = "SHIPPED"
)

var validTransitions = map[WMSStatus][]WMSStatus{
	WMSStatusReadyToPick: {WMSStatusPicking},
	WMSStatusPicking:     {WMSStatusPacked},
	WMSStatusPacked:      {WMSStatusShipped},
	WMSStatusShipped:     {}, // Terminal state
}

func (s WMSStatus) CanTransitionTo(target WMSStatus) bool {
	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

type Order struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	OrderSN               string                 `bun:"order_sn,pk" json:"order_sn"`
	ShopID                string                 `bun:"shop_id,notnull" json:"shop_id"`
	MarketplaceStatus     string                 `bun:"marketplace_status" json:"marketplace_status"`
	ShippingStatus        string                 `bun:"shipping_status" json:"shipping_status"`
	WMSStatus             WMSStatus              `bun:"wms_status" json:"wms_status"`
	TrackingNumber        string                 `bun:"tracking_number" json:"tracking_number,omitempty"`
	TotalAmount           float64                `bun:"total_amount" json:"total_amount"`
	RawMarketplacePayload map[string]interface{} `bun:"raw_marketplace_payload,type:jsonb" json:"raw_marketplace_payload,omitempty"`
	CreatedAt             time.Time              `bun:"created_at,default:current_timestamp" json:"created_at"`
	UpdatedAt             time.Time              `bun:"updated_at,default:current_timestamp" json:"updated_at"`

	// Relations
	Items []OrderItem `bun:"rel:has-many,join:order_sn=order_sn" json:"items,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	bun.BaseModel `bun:"table:order_items,alias:oi"`

	OrderSN  string  `bun:"order_sn,pk" json:"order_sn"`
	SKU      string  `bun:"sku,pk" json:"sku"`
	Quantity int     `bun:"quantity,notnull" json:"quantity"`
	Price    float64 `bun:"price,notnull" json:"price"`
}

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrOrderNotReadyToPick = errors.New("order is not ready to pick")
	ErrOrderNotPicking     = errors.New("order is not in picking status")
	ErrOrderNotPacked      = errors.New("order is not packed")
)

type WMSOrderListResponse struct {
	Orders []OrderSummary `json:"orders"`
}

type OrderSummary struct {
	OrderSN           string    `json:"order_sn"`
	WMSStatus         WMSStatus `json:"wms_status"`
	MarketplaceStatus string    `json:"marketplace_status"`
	ShippingStatus    string    `json:"shipping_status"`
	TrackingNumber    string    `json:"tracking_number,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (o *Order) ToSummary() OrderSummary {
	return OrderSummary{
		OrderSN:           o.OrderSN,
		WMSStatus:         o.WMSStatus,
		MarketplaceStatus: o.MarketplaceStatus,
		ShippingStatus:    o.ShippingStatus,
		TrackingNumber:    o.TrackingNumber,
		UpdatedAt:         o.UpdatedAt,
	}
}

type StatusUpdateResponse struct {
	OrderSN        string    `json:"order_sn"`
	WMSStatus      WMSStatus `json:"wms_status"`
	ShippingStatus string    `json:"shipping_status,omitempty"`
	TrackingNumber string    `json:"tracking_number,omitempty"`
}
