package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

type Config struct {
	Primary       Primary              `koanf:"primary" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis" validate:"required"`
	Integration   IntegrationConfig    `koanf:"integration" validate:"required"`
	Billing       BillingConfig        `koanf:"billing" validate:"required"`
	Observability *ObservabilityConfig `koanf:"observability"`
	Storage       StorageConfig        `koanf:"storage"`
}

type BillingConfig struct {
	Provider         string             `koanf:"provider" validate:"required"`
	MerchantCode     string             `koanf:"merchant_code" validate:"required"`
	SecureKey        string             `koanf:"secure_key" validate:"required"`
	BaseURL          string             `koanf:"base_url" validate:"required"`
	WebhookSecret    string             `koanf:"webhook_secret" validate:"required"`
	SuccessReturnURL string             `koanf:"success_return_url" validate:"required"`
	CancelReturnURL  string             `koanf:"cancel_return_url" validate:"required"`
	WebhookURL       string             `koanf:"webhook_url" validate:"required"`
	PlanPrices       map[string]float64 `koanf:"plan_prices" validate:"required"`

	ChargeExpiryMinutes    int `koanf:"charge_expiry_minutes" validate:"required"`
	DunningGraceDays       int `koanf:"dunning_grace_days" validate:"required"`
	DunningMaxRetries      int `koanf:"dunning_max_retries" validate:"required"`
	DunningRetryIntervalHr int `koanf:"dunning_retry_interval_hours" validate:"required"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}
type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

type StorageConfig struct {
	Provider string             `koanf:"provider"` // "r2" or "local"
	Local    LocalStorageConfig `koanf:"local"`
	R2       R2Config           `koanf:"r2"`
}

type R2Config struct {
	AccountID      string `koanf:"account_id"`
	AccessKeyID    string `koanf:"access_key_id"`
	SecretAccessKey string `koanf:"secret_access_key"`
	BucketName     string `koanf:"bucket_name"`
	PublicURL      string `koanf:"public_url"`
	PresignExpiry  int    `koanf:"presign_expiry"` // minutes, default 15
	KeyPrefix      string `koanf:"key_prefix"`     // e.g. "payments"
}

type LocalStorageConfig struct {
	BasePath string `koanf:"base_path"`
}

func LoadConfig() (*Config, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	k := koanf.New(".")

	err := k.Load(env.Provider("MONEYFLOW_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "MONEYFLOW_"))
	}), nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load initial env variables")
	}

	mainConfig := &Config{}

	err = k.Unmarshal("", mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal main config")
	}

	validate := validator.New()

	err = validate.Struct(mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("config validation failed")
	}

	// Set default observability config if not provided
	if mainConfig.Observability == nil {
		mainConfig.Observability = DefaultObservabilityConfig()
	}

	// Override service name and environment from primary config
	mainConfig.Observability.ServiceName = "moneyflow"
	mainConfig.Observability.Environment = mainConfig.Primary.Env

	// Validate observability config
	if err := mainConfig.Observability.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("invalid observability config")
	}

	return mainConfig, nil
}
