package server

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/integration/marketplace"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
)

type FiberServer struct {
	*fiber.App

	cfg *config.Config
	db  *database.DB

	// Handlers
	authHandler     *handler.AuthHandler
	orderHandler    *handler.OrderHandler
	wmsOrderHandler *handler.WMSOrderHandler

	// Middleware
	authMiddleware *middleware.AuthMiddleware
}

func New() *FiberServer {
	cfg := config.Load()

	db := database.NewBunDB(&cfg.Database)

	if err := db.Migrate(context.Background()); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	mpClient := marketplace.New(cfg)

	_, err := mpClient.Authorize(context.Background())
	if err != nil {
		log.Printf("marketplace authorize failed: %v", err)
	} else {
		if _, err := mpClient.ExchangeToken(context.Background()); err != nil {
			log.Printf("marketplace exchange token failed: %v", err)
		}
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	wmsOrderRepo := repository.NewWMSOrderRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, &cfg.JWT)
	orderService := service.NewOrderService(mpClient)
	wmsOrderService := service.NewWMSOrderService(wmsOrderRepo, mpClient)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	orderHandler := handler.NewOrderHandler(orderService)
	wmsOrderHandler := handler.NewWMSOrderHandler(wmsOrderService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "WMS Backend API",
	})
	// create worker 
	worker := marketplace.MarketPlaceSync{
		Client: mpClient,
		Repo: wmsOrderRepo,
	}

	go worker.SyncAllOrders(context.Background())

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("[Scheduler] Running marketplace order sync...")
			_, err := worker.SyncAllOrders(context.Background())
			if err != nil {
				log.Printf("[Scheduler] Sync failed: %v", err)
			}
		}
	}()

	return &FiberServer{
		App:             app,
		cfg:             cfg,
		db:              db,
		authHandler:     authHandler,
		orderHandler:    orderHandler,
		wmsOrderHandler: wmsOrderHandler,
		authMiddleware:  authMiddleware,
	}
}

func (s *FiberServer) ShutdownWithContext(ctx context.Context) error {
	s.db.Close()
	return s.App.ShutdownWithContext(ctx)
}
