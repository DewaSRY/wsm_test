package server

import (
	"backend/internal/domain"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *FiberServer) RegisterFiberRoutes() {
	// Global middleware
	s.App.Use(middleware.Recover())
	s.App.Use(middleware.Logger())

	// CORS middleware
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-Access-Token",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check endpoints
	s.App.Get("/", s.HelloWorldHandler)
	s.App.Get("/health", s.healthHandler)

	// API routes
	api := s.App.Group("/api")

	// Auth routes (public - no authentication required)
	auth := api.Group("/auth")
	auth.Post("/register", s.authHandler.Register)
	auth.Post("/login", s.authHandler.Login)
	auth.Post("/refresh", s.authHandler.RefreshToken)

	// Protected routes (require JWT authentication)
	protected := api.Group("", s.authMiddleware.JWTAuth())

	// User profile
	protected.Get("/me", s.authHandler.GetMe)

	// Order routes (protected) - Marketplace orders
	orders := protected.Group("/orders")
	orders.Get("/", s.orderHandler.ListOrders)
	orders.Get("/:order_sn", s.orderHandler.GetOrderDetail)

	// Warehouse routes (protected) - WMS orders
	warehouse := protected.Group("/warehouse")
	wmsOrders := warehouse.Group("/orders")
	wmsOrders.Get("/", s.wmsOrderHandler.ListOrders)
	wmsOrders.Get("/:order_sn", s.wmsOrderHandler.GetOrderDetail)
	wmsOrders.Post("/:order_sn/pick", s.wmsOrderHandler.PickOrder)
	wmsOrders.Post("/:order_sn/pack", s.wmsOrderHandler.PackOrder)
	wmsOrders.Post("/:order_sn/ship", s.wmsOrderHandler.ShipOrder)
}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	return c.JSON(domain.SuccessResponse("WMS Backend API", fiber.Map{
		"version": "1.0.0",
		"status":  "running",
	}))
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
