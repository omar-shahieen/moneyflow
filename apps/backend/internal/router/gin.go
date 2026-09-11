package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/middleware"
	"github.com/rs/zerolog"
)

type GinRouter struct {
	engine *gin.Engine
	config *config.Config
	db     *pgxpool.Pool
	logger zerolog.Logger
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

func (r *GinRouter) setupMiddleware() {
	r.engine.Use(middleware.RecoveryMiddleware(&r.logger))
	r.engine.Use(middleware.RequestIDMiddleware())
	r.engine.Use(middleware.CORSMiddleware(r.config.Server.CORSAllowedOrigins))
	r.engine.Use(middleware.SecureMiddleware())
	r.engine.Use(middleware.ContextEnrichmentMiddleware(r.logger))
	r.engine.Use(middleware.LoggerMiddleware(r.logger))
}

func (r *GinRouter) setupRoutes() {
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/status", r.healthCheck)

	v1 := r.engine.Group("/api/v1")
	_ = v1
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
