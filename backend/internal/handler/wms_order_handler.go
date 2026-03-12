package handler

import (
	"errors"
	"log"

	"backend/internal/domain"
	"backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type WMSOrderHandler struct {
	service *service.WMSOrderService
}

func NewWMSOrderHandler(service *service.WMSOrderService) *WMSOrderHandler {
	return &WMSOrderHandler{service: service}
}

// ListOrders handles GET /warehouse/orders
// @Summary List warehouse orders
// @Description Returns all orders with optional filtering by wms_status, sorted by updated_at descending
// @Tags Warehouse Orders
// @Accept json
// @Produce json
// @Param wms_status query string false "Filter by WMS status (READY_TO_PICK, PICKING, PACKED, SHIPPED)"
// @Success 200 {object} domain.WMSOrderListResponse
// @Failure 500 {object} domain.APIResponse
// @Router /warehouse/orders [get]
func (h *WMSOrderHandler) ListOrders(c *fiber.Ctx) error {
	var wmsStatus *domain.WMSStatus

	// Parse optional wms_status filter
	if statusStr := c.Query("wms_status"); statusStr != "" {
		status := domain.WMSStatus(statusStr)
		// Validate status
		if status != domain.WMSStatusReadyToPick &&
			status != domain.WMSStatusPicking &&
			status != domain.WMSStatusPacked &&
			status != domain.WMSStatusShipped {
			return c.Status(fiber.StatusBadRequest).JSON(
				domain.ErrorResponse("Bad Request", "Invalid wms_status value"),
			)
		}
		wmsStatus = &status
	}

	orders, err := h.service.ListOrders(c.Context(), wmsStatus)
	if err != nil {
		log.Printf("[WMSOrderHandler] Failed to list orders: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", "Failed to retrieve orders"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(domain.WMSOrderListResponse{
		Orders: orders,
	})
}

// GetOrderDetail handles GET /warehouse/orders/:order_sn
// @Summary Get order details
// @Description Returns full order details including items
// @Tags Warehouse Orders
// @Accept json
// @Produce json
// @Param order_sn path string true "Order SN"
// @Success 200 {object} domain.Order
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /warehouse/orders/{order_sn} [get]
func (h *WMSOrderHandler) GetOrderDetail(c *fiber.Ctx) error {
	orderSN := c.Params("order_sn")
	if orderSN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Order SN is required"),
		)
	}

	order, err := h.service.GetOrderDetail(c.Context(), orderSN)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				domain.ErrorResponse("Not Found", "Order not found"),
			)
		}
		log.Printf("[WMSOrderHandler] Failed to get order %s: %v", orderSN, err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", "Failed to retrieve order"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(order)
}

// PickOrder handles POST /warehouse/orders/:order_sn/pick
// @Summary Pick an order
// @Description Marks an order as being picked (READY_TO_PICK -> PICKING)
// @Tags Warehouse Orders
// @Accept json
// @Produce json
// @Param order_sn path string true "Order SN"
// @Success 200 {object} domain.StatusUpdateResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /warehouse/orders/{order_sn}/pick [post]
func (h *WMSOrderHandler) PickOrder(c *fiber.Ctx) error {
	orderSN := c.Params("order_sn")
	if orderSN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Order SN is required"),
		)
	}

	result, err := h.service.PickOrder(c.Context(), orderSN)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				domain.ErrorResponse("Not Found", "Order not found"),
			)
		}
		if errors.Is(err, domain.ErrOrderNotReadyToPick) {
			return c.Status(fiber.StatusBadRequest).JSON(
				domain.ErrorResponse("Bad Request", "Order is not ready to pick (must be READY_TO_PICK)"),
			)
		}
		log.Printf("[WMSOrderHandler] Failed to pick order %s: %v", orderSN, err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", "Failed to pick order"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// PackOrder handles POST /warehouse/orders/:order_sn/pack
// @Summary Pack an order
// @Description Marks an order as packed (PICKING -> PACKED)
// @Tags Warehouse Orders
// @Accept json
// @Produce json
// @Param order_sn path string true "Order SN"
// @Success 200 {object} domain.StatusUpdateResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /warehouse/orders/{order_sn}/pack [post]
func (h *WMSOrderHandler) PackOrder(c *fiber.Ctx) error {
	orderSN := c.Params("order_sn")
	if orderSN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Order SN is required"),
		)
	}

	result, err := h.service.PackOrder(c.Context(), orderSN)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				domain.ErrorResponse("Not Found", "Order not found"),
			)
		}
		if errors.Is(err, domain.ErrOrderNotPicking) {
			return c.Status(fiber.StatusBadRequest).JSON(
				domain.ErrorResponse("Bad Request", "Order is not in picking status (must be PICKING)"),
			)
		}
		log.Printf("[WMSOrderHandler] Failed to pack order %s: %v", orderSN, err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", "Failed to pack order"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// ShipOrder handles POST /warehouse/orders/:order_sn/ship
// @Summary Ship an order
// @Description Ships an order by calling marketplace API, receives tracking number, and syncs result (PACKED -> SHIPPED)
// @Tags Warehouse Orders
// @Accept json
// @Produce json
// @Param order_sn path string true "Order SN"
// @Param body body ShipRequest true "Ship request body"
// @Success 200 {object} domain.StatusUpdateResponse
// @Failure 400 {object} domain.APIResponse
// @Failure 404 {object} domain.APIResponse
// @Failure 500 {object} domain.APIResponse
// @Router /warehouse/orders/{order_sn}/ship [post]
func (h *WMSOrderHandler) ShipOrder(c *fiber.Ctx) error {
	orderSN := c.Params("order_sn")
	if orderSN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Order SN is required"),
		)
	}

	var req ShipRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Invalid request body"),
		)
	}

	if req.ChannelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "channel_id is required"),
		)
	}

	// Get access token from Authorization header or request body
	accessToken := req.AccessToken
	if accessToken == "" {
		accessToken = c.Get("X-Access-Token")
	}
	if accessToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Access token is required"),
		)
	}

	result, err := h.service.ShipOrder(c.Context(), orderSN, req.ChannelID, accessToken)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				domain.ErrorResponse("Not Found", "Order not found"),
			)
		}
		if errors.Is(err, domain.ErrOrderNotPacked) {
			return c.Status(fiber.StatusBadRequest).JSON(
				domain.ErrorResponse("Bad Request", "Order is not packed (must be PACKED)"),
			)
		}
		log.Printf("[WMSOrderHandler] Failed to ship order %s: %v", orderSN, err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", "Failed to ship order: "+err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// ShipRequest represents the request body for shipping an order
type ShipRequest struct {
	ChannelID   string `json:"channel_id"`
	AccessToken string `json:"access_token,omitempty"`
}
