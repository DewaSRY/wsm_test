package handler

import (
	"backend/internal/domain"
	"backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// ListOrders returns the list of orders from the marketplace
func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	orders, err := h.orderService.ListOrders(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		domain.SuccessResponse("Orders retrieved successfully", orders),
	)
}

// GetOrderDetail returns the details of a specific order
func (h *OrderHandler) GetOrderDetail(c *fiber.Ctx) error {
	orderSN := c.Params("order_sn")
	if orderSN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Order SN is required"),
		)
	}

	order, err := h.orderService.GetOrderDetail(c.Context(), orderSN)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		domain.SuccessResponse("Order retrieved successfully", order),
	)
}

