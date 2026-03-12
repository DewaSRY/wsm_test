package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"backend/internal/config"
	"backend/internal/domain"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type DB struct {
	*bun.DB
}

var dbInstance *DB

func NewBunDB(cfg *config.DatabaseConfig) *DB {
	if dbInstance != nil {
		return dbInstance
	}

	dsn := cfg.DSN()
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	dbInstance = &DB{DB: db}
	return dbInstance
}

func (db *DB) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "Database is healthy"

	dbStats := db.DB.DB.Stats()
	stats["open_connections"] = fmt.Sprintf("%d", dbStats.OpenConnections)
	stats["in_use"] = fmt.Sprintf("%d", dbStats.InUse)
	stats["idle"] = fmt.Sprintf("%d", dbStats.Idle)

	return stats
}

func (db *DB) Close() error {
	return db.DB.Close()
}

// Migrate creates or updates database tables based on domain models
func (db *DB) Migrate(ctx context.Context) error {
	// Register all models that need tables
	models := []interface{}{
		(*domain.User)(nil),
		(*domain.Order)(nil),
		(*domain.OrderItem)(nil),
	}

	log.Println("[Database] Running auto-migration...")

	for _, model := range models {
		// Create table if not exists
		_, err := db.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create table for %T: %w", model, err)
		}
		log.Printf("[Database] Migrated table for %T", model)
	}

	// Create indexes for orders table
	db.createIndexIfNotExists(ctx, "orders", "idx_orders_wms_status", "wms_status")
	db.createIndexIfNotExists(ctx, "orders", "idx_orders_shop_id", "shop_id")
	db.createIndexIfNotExists(ctx, "orders", "idx_orders_updated_at", "updated_at DESC")

	// Create indexes for order_items table
	db.createIndexIfNotExists(ctx, "order_items", "idx_order_items_sku", "sku")

	log.Println("[Database] Auto-migration completed")
	return nil
}

// createIndexIfNotExists creates an index if it doesn't already exist
func (db *DB) createIndexIfNotExists(ctx context.Context, table, indexName, columns string) {
	query := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (%s)`, indexName, table, columns)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("[Database] Warning: failed to create index %s: %v", indexName, err)
	}
}


