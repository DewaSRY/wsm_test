package service

import (
	"backend/internal/domain"
	"backend/internal/integration/marketplace"
	"context"
)

type OrderService struct {
	marketplaceClient *marketplace.MarketplaceClient
}

func NewOrderService(marketplaceClient *marketplace.MarketplaceClient) *OrderService {
	return &OrderService{
		marketplaceClient: marketplaceClient,
	}
}

func (s *OrderService) ListOrders(ctx context.Context) ([]domain.MarketplaceOrder, error) {
	return s.marketplaceClient.ListOrders(ctx)
}

func (s *OrderService) GetOrderDetail(ctx context.Context, orderSN string) (*domain.MarketplaceOrder, error) {
	return s.marketplaceClient.GetOrderDetail(ctx, orderSN)
}
