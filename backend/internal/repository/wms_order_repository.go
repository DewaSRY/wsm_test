package repository

import (
	"context"

	"backend/internal/database"
	"backend/internal/domain"
)

type WMSOrderRepository struct {
	db *database.DB
}

func NewWMSOrderRepository(db *database.DB) *WMSOrderRepository {
	return &WMSOrderRepository{db: db}
}

// List returns all orders with optional filtering by wms_status
func (r *WMSOrderRepository) List(ctx context.Context, wmsStatus *domain.WMSStatus) ([]domain.Order, error) {
	var orders []domain.Order

	query := r.db.NewSelect().
		Model(&orders).
		Order("updated_at DESC")

	if wmsStatus != nil {
		query = query.Where("wms_status = ?", *wmsStatus)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// GetByOrderSN returns an order by its order_sn with items
func (r *WMSOrderRepository) GetByOrderSN(ctx context.Context, orderSN string) (*domain.Order, error) {
	order := new(domain.Order)

	err := r.db.NewSelect().
		Model(order).
		Where("order_sn = ?", orderSN).
		Relation("Items").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return order, nil
}

// UpdateStatus updates the WMS status of an order
func (r *WMSOrderRepository) UpdateStatus(ctx context.Context, orderSN string, status domain.WMSStatus) error {
	_, err := r.db.NewUpdate().
		Model((*domain.Order)(nil)).
		Set("wms_status = ?", status).
		Where("order_sn = ?", orderSN).
		Exec(ctx)

	return err
}

// UpdateShipping updates shipping information after marketplace ship call
func (r *WMSOrderRepository) UpdateShipping(ctx context.Context, orderSN string, trackingNumber string, shippingStatus string, wmsStatus domain.WMSStatus) error {
	_, err := r.db.NewUpdate().
		Model((*domain.Order)(nil)).
		Set("tracking_number = ?", trackingNumber).
		Set("shipping_status = ?", shippingStatus).
		Set("wms_status = ?", wmsStatus).
		Where("order_sn = ?", orderSN).
		Exec(ctx)

	return err
}

// UpdateMarketplaceStatus updates marketplace and shipping status (from marketplace sync)
func (r *WMSOrderRepository) UpdateMarketplaceStatus(ctx context.Context, orderSN string, marketplaceStatus string, shippingStatus string) error {
	_, err := r.db.NewUpdate().
		Model((*domain.Order)(nil)).
		Set("marketplace_status = ?", marketplaceStatus).
		Set("shipping_status = ?", shippingStatus).
		Where("order_sn = ?", orderSN).
		Exec(ctx)

	return err
}

// Create creates a new order with items (for ingesting from marketplace)
func (r *WMSOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert order
	_, err = tx.NewInsert().Model(order).Exec(ctx)
	if err != nil {
		return err
	}

	// Insert items if any
	if len(order.Items) > 0 {
		for i := range order.Items {
			order.Items[i].OrderSN = order.OrderSN
		}
		_, err = tx.NewInsert().Model(&order.Items).Exec(ctx)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

// Exists checks if an order exists
func (r *WMSOrderRepository) Exists(ctx context.Context, orderSN string) (bool, error) {
	return r.db.NewSelect().
		Model((*domain.Order)(nil)).
		Where("order_sn = ?", orderSN).
		Exists(ctx)
}
