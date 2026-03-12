package marketplace

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"log"
)

type MarketPlaceSync struct {
	Client *MarketplaceClient
	Repo   *repository.WMSOrderRepository
}

func NewMarketPlaceSync(client *MarketplaceClient, repo *repository.WMSOrderRepository) *MarketPlaceSync {
	return &MarketPlaceSync{
		Client: client,
		Repo:   repo,
	}
}

type SyncResult struct {
	TotalFetched int      `json:"total_fetched"`
	Created      int      `json:"created"`
	Updated      int      `json:"updated"`
	Skipped      int      `json:"skipped"`
	Errors       []string `json:"errors,omitempty"`
}

func (s *MarketPlaceSync) SyncAllOrders(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{}

	log.Println("[MarketPlaceSync] Fetching orders from marketplace...")
	marketplaceOrders, err := s.Client.ListOrders(ctx)
	if err != nil {
		log.Printf("[MarketPlaceSync] Failed to fetch orders from marketplace: %v", err)
		return nil, err
	}

	result.TotalFetched = len(marketplaceOrders)
	log.Printf("[MarketPlaceSync] Fetched %d orders from marketplace", result.TotalFetched)

	for _, mpOrder := range marketplaceOrders {
		err := s.syncSingleOrder(ctx, &mpOrder, result)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			log.Printf("[MarketPlaceSync] Error syncing order %s: %v", mpOrder.OrderSN, err)
		}
	}

	log.Printf("[MarketPlaceSync] Sync completed: created=%d, updated=%d, skipped=%d, errors=%d",
		result.Created, result.Updated, result.Skipped, len(result.Errors))

	return result, nil
}

func (s *MarketPlaceSync) syncSingleOrder(ctx context.Context, mpOrder *domain.MarketplaceOrder, result *SyncResult) error {
	exists, err := s.Repo.Exists(ctx, mpOrder.OrderSN)
	if err != nil {
		return err
	}

	if exists {
		existingOrder, err := s.Repo.GetByOrderSN(ctx, mpOrder.OrderSN)
		if err != nil {
			return err
		}

		if existingOrder.MarketplaceStatus != mpOrder.Status && existingOrder.WMSStatus != domain.WMSStatusShipped {
			err = s.Repo.UpdateMarketplaceStatus(ctx, mpOrder.OrderSN, mpOrder.Status, mpOrder.ShippingStatus)
			if err != nil {
				return err
			}
			result.Updated++
			log.Printf("[MarketPlaceSync] Updated order %s marketplace status: %s -> %s",
				mpOrder.OrderSN, existingOrder.MarketplaceStatus, mpOrder.Status)
		} else {
			result.Skipped++
		}
		return nil
	}

	order := s.convertToLocalOrder(mpOrder)
	err = s.Repo.Create(ctx, order)
	if err != nil {
		return err
	}

	result.Created++
	log.Printf("[MarketPlaceSync] Created new order %s with status %s", order.OrderSN, order.WMSStatus)
	return nil
}

func (s *MarketPlaceSync) convertToLocalOrder(mpOrder *domain.MarketplaceOrder) *domain.Order {
	// Convert items
	items := make([]domain.OrderItem, len(mpOrder.Items))
	for i, item := range mpOrder.Items {
		items[i] = domain.OrderItem{
			OrderSN:  mpOrder.OrderSN,
			SKU:      item.SKU,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	return &domain.Order{
		OrderSN:           mpOrder.OrderSN,
		ShopID:            mpOrder.ShopID,
		MarketplaceStatus: mpOrder.Status,
		ShippingStatus:    mpOrder.ShippingStatus,
		WMSStatus:         domain.WMSStatusReadyToPick, 
		TotalAmount:       mpOrder.TotalAmount,
		Items:             items,
	}
}