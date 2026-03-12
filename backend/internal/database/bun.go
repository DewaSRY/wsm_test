package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"backend/internal/config"

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

func (db *DB) Migrate(ctx context.Context) error {
	models := []interface{}{
		// (*domain.Merchant)(nil),
		// (*domain.OAuthToken)(nil),
		// (*domain.Order)(nil),
		// (*domain.OrderItem)(nil),
		// (*domain.Shipment)(nil),
		// (*domain.WebhookEvent)(nil),
	}

	for _, model := range models {
		_, err := db.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create table for %T: %w", model, err)
		}
	}



	return nil
}


