package router

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/config"
	"github.com/user/small-erp/internal/handler"
	"github.com/user/small-erp/internal/middleware"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// CORS harus paling awal, sebelum middleware lain.
	r.Use(middleware.CORS([]string{"http://localhost:5173"}))

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(chiMiddleware.RealIP)

	// Health check (public)
	healthHandler := handler.NewHealthHandler(db)
	r.Get("/health", healthHandler.Health)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		authHandler := handler.NewAuthHandler(db, cfg)
		r.Post("/auth/login", authHandler.Login)

		sessionHandler := handler.NewSessionHandler(db, cfg)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", sessionHandler.Refresh)
		r.Post("/auth/logout", sessionHandler.Logout)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))

			r.Use(middleware.Audit(db, cfg.JWTSecret))

			auditHandler := handler.NewAuditHandler(db)
			r.Get("/audit-logs", middleware.RequirePermission("audit:read", auditHandler.List))

			// User management
			userHandler := handler.NewUserHandler(db)
			r.Get("/users", middleware.RequirePermission("users:read", userHandler.List))
			r.Post("/users", middleware.RequirePermission("users:create", userHandler.Create))
			r.Get("/users/{id}", middleware.RequirePermission("users:read", userHandler.GetByID))
			r.Put("/users/{id}", middleware.RequirePermission("users:update", userHandler.Update))
			r.Delete("/users/{id}", middleware.RequirePermission("users:delete", userHandler.Delete))

			// Role management
			roleHandler := handler.NewRoleHandler(db)
			r.Get("/roles", middleware.RequirePermission("roles:read", roleHandler.List))
			r.Post("/roles", middleware.RequirePermission("roles:create", roleHandler.Create))
			r.Get("/roles/{id}", middleware.RequirePermission("roles:read", roleHandler.GetByID))
			r.Put("/roles/{id}", middleware.RequirePermission("roles:update", roleHandler.Update))
			r.Delete("/roles/{id}", middleware.RequirePermission("roles:delete", roleHandler.Delete))

			// Product Category management
			categoryHandler := handler.NewCategoryHandler(db)
			r.Get("/product-categories", middleware.RequirePermission("products:read", categoryHandler.List))
			r.Post("/product-categories", middleware.RequirePermission("products:create", categoryHandler.Create))
			r.Get("/product-categories/{id}", middleware.RequirePermission("products:read", categoryHandler.GetByID))
			r.Put("/product-categories/{id}", middleware.RequirePermission("products:update", categoryHandler.Update))
			r.Delete("/product-categories/{id}", middleware.RequirePermission("products:delete", categoryHandler.Delete))

			// Unit management
			unitHandler := handler.NewUnitHandler(db)
			r.Get("/units", middleware.RequirePermission("products:read", unitHandler.List))
			r.Post("/units", middleware.RequirePermission("products:create", unitHandler.Create))
			r.Get("/units/{id}", middleware.RequirePermission("products:read", unitHandler.GetByID))
			r.Put("/units/{id}", middleware.RequirePermission("products:update", unitHandler.Update))
			r.Delete("/units/{id}", middleware.RequirePermission("products:delete", unitHandler.Delete))

			// Product management
			productHandler := handler.NewProductHandler(db)
			r.Get("/products", middleware.RequirePermission("products:read", productHandler.List))
			r.Post("/products", middleware.RequirePermission("products:create", productHandler.Create))
			r.Get("/products/{id}", middleware.RequirePermission("products:read", productHandler.GetByID))
			r.Put("/products/{id}", middleware.RequirePermission("products:update", productHandler.Update))
			r.Delete("/products/{id}", middleware.RequirePermission("products:delete", productHandler.Delete))

			// Customer management
			customerHandler := handler.NewCustomerHandler(db)
			r.Get("/customers", middleware.RequirePermission("customers:read", customerHandler.List))
			r.Post("/customers", middleware.RequirePermission("customers:create", customerHandler.Create))
			r.Get("/customers/{id}", middleware.RequirePermission("customers:read", customerHandler.GetByID))
			r.Put("/customers/{id}", middleware.RequirePermission("customers:update", customerHandler.Update))
			r.Delete("/customers/{id}", middleware.RequirePermission("customers:delete", customerHandler.Delete))

			// Supplier management
			supplierHandler := handler.NewSupplierHandler(db)
			r.Get("/suppliers", middleware.RequirePermission("suppliers:read", supplierHandler.List))
			r.Post("/suppliers", middleware.RequirePermission("suppliers:create", supplierHandler.Create))
			r.Get("/suppliers/{id}", middleware.RequirePermission("suppliers:read", supplierHandler.GetByID))
			r.Put("/suppliers/{id}", middleware.RequirePermission("suppliers:update", supplierHandler.Update))
			r.Delete("/suppliers/{id}", middleware.RequirePermission("suppliers:delete", supplierHandler.Delete))

			// Purchase Order management
			poHandler := handler.NewPurchaseOrderHandler(db)
			r.Get("/purchase-orders", middleware.RequirePermission("purchases:read", poHandler.List))
			r.Post("/purchase-orders", middleware.RequirePermission("purchases:create", poHandler.Create))
			r.Get("/purchase-orders/{id}", middleware.RequirePermission("purchases:read", poHandler.GetByID))
			r.Put("/purchase-orders/{id}", middleware.RequirePermission("purchases:update", poHandler.Update))
			r.Patch("/purchase-orders/{id}/status", middleware.RequirePermission("purchases:update", poHandler.UpdateStatus))

			// Goods Receipt
			grHandler := handler.NewGoodsReceiptHandler(db)
			r.Get("/goods-receipts", middleware.RequirePermission("purchases:read", grHandler.List))
			r.Post("/goods-receipts", middleware.RequirePermission("purchases:create", grHandler.Create))
			r.Get("/goods-receipts/{id}", middleware.RequirePermission("purchases:read", grHandler.GetByID))

			// Inventory management
			inventoryHandler := handler.NewInventoryHandler(db)
			r.Get("/inventory", middleware.RequirePermission("inventory:read", inventoryHandler.List))
			r.Get("/inventory/movements", middleware.RequirePermission("inventory:read", inventoryHandler.Movements))
			r.Post("/inventory/adjustments", middleware.RequirePermission("inventory:create", inventoryHandler.Adjustment))

			// Sales Order management
			soHandler := handler.NewSalesOrderHandler(db)
			r.Get("/sales-orders", middleware.RequirePermission("sales:read", soHandler.List))
			r.Post("/sales-orders", middleware.RequirePermission("sales:create", soHandler.Create))
			r.Get("/sales-orders/{id}", middleware.RequirePermission("sales:read", soHandler.GetByID))
			r.Put("/sales-orders/{id}", middleware.RequirePermission("sales:update", soHandler.Update))
			r.Patch("/sales-orders/{id}/status", middleware.RequirePermission("sales:update", soHandler.UpdateStatus))

			// Finance - Chart of Accounts
			financeHandler := handler.NewFinanceHandler(db)
			r.Get("/accounts", middleware.RequirePermission("finance:read", financeHandler.ListAccounts))
			r.Post("/accounts", middleware.RequirePermission("finance:create", financeHandler.CreateAccount))
			r.Get("/accounts/{id}", middleware.RequirePermission("finance:read", financeHandler.GetAccount))
			r.Put("/accounts/{id}", middleware.RequirePermission("finance:update", financeHandler.UpdateAccount))

			// Journal entries
			r.Get("/journal-entries", middleware.RequirePermission("finance:read", financeHandler.ListJournalEntries))
			r.Post("/journal-entries", middleware.RequirePermission("finance:create", financeHandler.CreateJournalEntry))
			r.Get("/journal-entries/{id}", middleware.RequirePermission("finance:read", financeHandler.GetJournalEntry))

			// Payments
			paymentHandler := handler.NewPaymentHandler(db)
			r.Get("/payments", middleware.RequirePermission("finance:read", paymentHandler.List))
			r.Post("/payments", middleware.RequirePermission("finance:create", paymentHandler.Create))
			r.Get("/payments/{id}", middleware.RequirePermission("finance:read", paymentHandler.GetByID))

			// Accounts Payable / Receivable
			r.Get("/accounts-payable", middleware.RequirePermission("finance:read", financeHandler.ListAP))
			r.Get("/accounts-receivable", middleware.RequirePermission("finance:read", financeHandler.ListAR))

			// Reports
			reportHandler := handler.NewReportHandler(db)
			r.Get("/reports/sales", middleware.RequirePermission("reports:read", reportHandler.SalesReport))
			r.Get("/reports/purchases", middleware.RequirePermission("reports:read", reportHandler.PurchaseReport))
			r.Get("/reports/inventory", middleware.RequirePermission("reports:read", reportHandler.InventoryReport))
			r.Get("/reports/stock-movements", middleware.RequirePermission("reports:read", reportHandler.StockMovementReport))
			r.Get("/reports/accounts-payable", middleware.RequirePermission("reports:read", reportHandler.APReport))
			r.Get("/reports/accounts-receivable", middleware.RequirePermission("reports:read", reportHandler.ARReport))
			r.Get("/reports/cash-flow", middleware.RequirePermission("reports:read", reportHandler.CashFlowReport))
			r.Get("/reports/profit-loss", middleware.RequirePermission("reports:read", reportHandler.ProfitLossReport))
			r.Get("/reports/balance-sheet", middleware.RequirePermission("reports:read", reportHandler.BalanceSheetReport))
		})
	})

	return r
}
