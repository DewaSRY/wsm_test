package server

import (
	"context"

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
	authHandler  *handler.AuthHandler
	orderHandler *handler.OrderHandler

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

	userRepo := repository.NewUserRepository(db)

	authService := service.NewAuthService(userRepo, &cfg.JWT)
	orderService := service.NewOrderService(mpClient)

	authHandler := handler.NewAuthHandler(authService)
	orderHandler := handler.NewOrderHandler(orderService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "WMS Backend API",
	})

	return &FiberServer{
		App:            app,
		cfg:            cfg,
		db:             db,
		authHandler:    authHandler,
		orderHandler:   orderHandler,
		authMiddleware: authMiddleware,
	}
}

func (s *FiberServer) ShutdownWithContext(ctx context.Context) error {
	s.db.Close()
	return s.App.ShutdownWithContext(ctx)
}
