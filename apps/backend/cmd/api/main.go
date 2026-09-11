package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/database"
	"github.com/omar-shahieen/moneyflow/internal/handler"
	"github.com/omar-shahieen/moneyflow/internal/logger"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/router"
	"github.com/omar-shahieen/moneyflow/internal/service"
	"github.com/redis/go-redis/v9"
)

const DefaultContextTimeout = 30

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	db, err := database.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})
	defer redisClient.Close()

	r := router.NewGinRouter(cfg, db.Pool, log)

	userRepo := repository.NewUserRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)
	categoryService := service.NewCategoryService(categoryRepo, userRepo, nil)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	r.SetHandlers(&router.Handlers{
		Category: categoryHandler,
	})

	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r.Engine(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		log.Info().
			Str("port", cfg.Server.Port).
			Str("env", cfg.Primary.Env).
			Msg("starting server")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close database connection")
	}

	stop()
	log.Info().Msg("server exited properly")
}
