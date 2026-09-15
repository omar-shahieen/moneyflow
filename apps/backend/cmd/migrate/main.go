package main

import (
	"context"

	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/database"
	"github.com/omar-shahieen/moneyflow/internal/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	if err := database.Migrate(context.Background(), &log, cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to run database migrations")
	}

	log.Info().Msg("migrations completed successfully")
}
