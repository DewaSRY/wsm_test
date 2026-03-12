package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"backend/internal/domain"
	"backend/internal/integration/marketplace"
	"backend/internal/repository"
)

type WMSOrderService struct {
	repo     *repository.WMSOrderRepository
	mpClient *marketplace.MarketplaceClient
}

func NewWMSOrderService(repo *repository.WMSOrderRepository, mpClient *marketplace.MarketplaceClient) *WMSOrderService {
	return &WMSOrderService{
		repo:     repo,
		mpClient: mpClient,
	}
}

func (s *WMSOrderService) ListOrders(ctx context.Context, wmsStatus *domain.WMSStatus) ([]domain.OrderSummary, error) {
	orders, err := s.repo.List(ctx, wmsStatus)
	if err != nil {
		log.Printf("[WMSOrderService] Failed to list orders: %v", err)
		return nil, err
	}

	summaries := make([]domain.OrderSummary, len(orders))
	for i, order := range orders {
		summaries[i] = order.ToSummary()
	}

	return summaries, nil
}

func (s *WMSOrderService) GetOrderDetail(ctx context.Context, orderSN string) (*domain.Order, error) {
	order, err := s.repo.GetByOrderSN(ctx, orderSN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		log.Printf("[WMSOrderService] Failed to get order %s: %v", orderSN, err)
		return nil, err
	}

	return order, nil
}

// PickOrder marks an order as being picked (READY_TO_PICK -> PICKING)
func (s *WMSOrderService) PickOrder(ctx context.Context, orderSN string) (*domain.StatusUpdateResponse, error) {
	order, err := s.repo.GetByOrderSN(ctx, orderSN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	// Validate current status
	if order.WMSStatus != domain.WMSStatusReadyToPick {
		return nil, domain.ErrOrderNotReadyToPick
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, orderSN, domain.WMSStatusPicking); err != nil {
		log.Printf("[WMSOrderService] Failed to update status for order %s: %v", orderSN, err)
		return nil, err
	}

	return &domain.StatusUpdateResponse{
		OrderSN:   orderSN,
		WMSStatus: domain.WMSStatusPicking,
	}, nil
}

// PackOrder marks an order as packed (PICKING -> PACKED)
func (s *WMSOrderService) PackOrder(ctx context.Context, orderSN string) (*domain.StatusUpdateResponse, error) {
	order, err := s.repo.GetByOrderSN(ctx, orderSN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	// Validate current status
	if order.WMSStatus != domain.WMSStatusPicking {
		return nil, domain.ErrOrderNotPicking
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, orderSN, domain.WMSStatusPacked); err != nil {
		log.Printf("[WMSOrderService] Failed to update status for order %s: %v", orderSN, err)
		return nil, err
	}

	return &domain.StatusUpdateResponse{
		OrderSN:   orderSN,
		WMSStatus: domain.WMSStatusPacked,
	}, nil
}

// ShipOrder ships an order by calling marketplace API and syncing the result (PACKED -> SHIPPED)
func (s *WMSOrderService) ShipOrder(ctx context.Context, orderSN string, channelID string ) (*domain.StatusUpdateResponse, error) {
	order, err := s.repo.GetByOrderSN(ctx, orderSN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	// Validate current status
	if order.WMSStatus != domain.WMSStatusPacked {
		return nil, domain.ErrOrderNotPacked
	}

	// Call marketplace API to ship the order
	shipResp, err := s.mpClient.ShipOrder(ctx, orderSN, channelID)
	if err != nil {
		log.Printf("[WMSOrderService] Failed to ship order %s via marketplace API: %v", orderSN, err)
		return nil, err
	}

	// Update order with shipping info
	if err := s.repo.UpdateShipping(ctx, orderSN, shipResp.TrackingNo, shipResp.ShippingStatus, domain.WMSStatusShipped); err != nil {
		log.Printf("[WMSOrderService] Failed to update shipping info for order %s: %v", orderSN, err)
		return nil, err
	}

	return &domain.StatusUpdateResponse{
		OrderSN:        orderSN,
		WMSStatus:      domain.WMSStatusShipped,
		ShippingStatus: shipResp.ShippingStatus,
		TrackingNumber: shipResp.TrackingNo,
	}, nil
}

// IngestOrder ingests an order from marketplace into WMS
func (s *WMSOrderService) IngestOrder(ctx context.Context, mpOrder *domain.MarketplaceOrder) error {
	// Check if order already exists
	exists, err := s.repo.Exists(ctx, mpOrder.OrderSN)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already ingested
	}

	// Calculate total amount
	var totalAmount float64
	items := make([]domain.OrderItem, len(mpOrder.Items))
	for i, item := range mpOrder.Items {
		totalAmount += item.Price * float64(item.Quantity)
		items[i] = domain.OrderItem{
			OrderSN:  mpOrder.OrderSN,
			SKU:      item.SKU,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	order := &domain.Order{
		OrderSN:           mpOrder.OrderSN,
		ShopID:            mpOrder.ShopID,
		MarketplaceStatus: mpOrder.Status,
		ShippingStatus:    mpOrder.ShippingStatus,
		WMSStatus:         domain.WMSStatusReadyToPick,
		TotalAmount:       totalAmount,
		Items:             items,
	}

	return s.repo.Create(ctx, order)
}
