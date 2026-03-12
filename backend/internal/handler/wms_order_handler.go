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


func (h *WMSOrderHandler) ListOrders(c *fiber.Ctx) error {
	var wmsStatus *domain.WMSStatus

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


	result, err := h.service.ShipOrder(c.Context(), orderSN, req.ChannelID)
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

type ShipRequest struct {
	ChannelID   string `json:"channel_id"`
	AccessToken string `json:"access_token,omitempty"`
}
