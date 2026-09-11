package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/handler"
	"github.com/omar-shahieen/moneyflow/internal/middleware"
	"github.com/omar-shahieen/moneyflow/internal/service"
	"github.com/rs/zerolog"
)

type GinRouter struct {
	engine   *gin.Engine
	config   *config.Config
	db       *pgxpool.Pool
	logger   zerolog.Logger
	handlers *Handlers
}

type Handlers struct {
	Category     *handler.CategoryHandler
	Transaction  *handler.TransactionHandler
	Budget       *handler.BudgetHandler
	Subscription *handler.SubscriptionHandler
	Recurring    *handler.RecurringRuleHandler
	Receipt      *handler.ReceiptHandler
	Import       *handler.ImportHandler
	Report       *handler.ReportHandler
}

type Services struct {
	Category *service.CategoryService
}

func NewGinRouter(cfg *config.Config, db *pgxpool.Pool, log zerolog.Logger) *GinRouter {
	if cfg.Primary.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	r := &GinRouter{
		engine: engine,
		config: cfg,
		db:     db,
		logger: log,
	}

	r.setupMiddleware()
	r.setupRoutes()

	return r
}

func (r *GinRouter) SetHandlers(h *Handlers) {
	r.handlers = h
	r.registerDomainRoutes()
}

func (r *GinRouter) setupMiddleware() {
	r.engine.Use(middleware.RecoveryMiddleware(&r.logger))
	r.engine.Use(middleware.RequestIDMiddleware())
	r.engine.Use(middleware.CORSMiddleware(r.config.Server.CORSAllowedOrigins))
	r.engine.Use(middleware.SecureMiddleware())
	r.engine.Use(middleware.ContextEnrichmentMiddleware(r.logger))
	r.engine.Use(middleware.ErrorHandlerMiddleware())
	r.engine.Use(middleware.LoggerMiddleware(r.logger))
}

func (r *GinRouter) setupRoutes() {
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/status", r.healthCheck)
}

func (r *GinRouter) registerDomainRoutes() {
	if r.handlers == nil {
		return
	}

	v1 := r.engine.Group("/api/v1")

	if r.handlers.Category != nil {
		categories := v1.Group("/categories")
		categories.GET("", r.handlers.Category.List)
		categories.GET("/:id", r.handlers.Category.GetByID)
		categories.POST("", r.handlers.Category.Create)
		categories.PATCH("/:id", r.handlers.Category.Update)
		categories.DELETE("/:id", r.handlers.Category.Delete)
	}

	if r.handlers.Transaction != nil {
		transactions := v1.Group("/transactions")
		transactions.GET("", r.handlers.Transaction.List)
		transactions.GET("/summary", r.handlers.Transaction.Summary)
		transactions.GET("/:id", r.handlers.Transaction.GetByID)
		transactions.POST("", r.handlers.Transaction.Create)
		transactions.PATCH("/:id", r.handlers.Transaction.Update)
		transactions.DELETE("/:id", r.handlers.Transaction.Delete)
	}

	if r.handlers.Budget != nil {
		budgets := v1.Group("/budgets")
		budgets.GET("", r.handlers.Budget.List)
		budgets.GET("/:id", r.handlers.Budget.GetByID)
		budgets.POST("", r.handlers.Budget.Create)
		budgets.PATCH("/:id", r.handlers.Budget.Update)
		budgets.POST("/:id/members", r.handlers.Budget.AddMember)
		budgets.DELETE("/:id/members/:userId", r.handlers.Budget.RemoveMember)
	}

	if r.handlers.Subscription != nil {
		subscription := v1.Group("/subscription")
		subscription.GET("", r.handlers.Subscription.Get)
	}

	if r.handlers.Recurring != nil {
		recurring := v1.Group("/recurring-rules")
		recurring.GET("", r.handlers.Recurring.List)
		recurring.GET("/:id", r.handlers.Recurring.GetByID)
		recurring.POST("", r.handlers.Recurring.Create)
		recurring.PATCH("/:id", r.handlers.Recurring.Update)
		recurring.DELETE("/:id", r.handlers.Recurring.Delete)
	}

	if r.handlers.Receipt != nil {
		transactions := v1.Group("/transactions")
		transactions.POST("/:id/receipt", r.handlers.Receipt.GetUploadURL)
	}

	if r.handlers.Import != nil {
		imports := v1.Group("/imports")
		imports.POST("", r.handlers.Import.Create)
		imports.GET("/:id", r.handlers.Import.GetByID)
	}

	if r.handlers.Report != nil {
		reports := v1.Group("/reports")
		reports.GET("", r.handlers.Report.List)
		reports.GET("/:id", r.handlers.Report.GetByID)
		reports.POST("", r.handlers.Report.Create)
	}
}

func (r *GinRouter) healthCheck(c *gin.Context) {
	response := gin.H{
		"status":      "healthy",
		"timestamp":   time.Now().UTC(),
		"environment": r.config.Primary.Env,
		"checks":      gin.H{},
	}

	checks := response["checks"].(gin.H)
	isHealthy := true

	if r.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dbStart := time.Now()
		if err := r.db.Ping(ctx); err != nil {
			checks["database"] = gin.H{
				"status":        "unhealthy",
				"response_time": time.Since(dbStart).String(),
				"error":         err.Error(),
			}
			isHealthy = false
		} else {
			checks["database"] = gin.H{
				"status":        "healthy",
				"response_time": time.Since(dbStart).String(),
			}
		}
	}

	if !isHealthy {
		response["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (r *GinRouter) Engine() *gin.Engine {
	return r.engine
}
