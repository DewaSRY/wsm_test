package domain

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

type Product struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type ProductResponse struct {
	Message string  `json:"message"`
	Data    Product `json:"data"`
}


type LogisticChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type ShipOrderRequest struct {
	OrderSN   string `json:"order_sn" validate:"required"`
	ChannelID string `json:"channel_id" validate:"required"`
}

type ShipOrderResponse struct {
	OrderSN        string `json:"order_sn"`
	TrackingNo     string `json:"tracking_no"`
	ShippingStatus string `json:"shipping_status"`
}

type LogisticChannelsResponse struct {
	Message string            `json:"message"`
	Data    []LogisticChannel `json:"data"`
}
