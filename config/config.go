package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	NATS     NATSConfig     `mapstructure:"nats"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	OTel     OTelConfig     `mapstructure:"otel"`

	CorsOrigins string `mapstructure:"cors__origins"`

	// Service discovery
	UsersService    ServiceConfig `mapstructure:"users_service"`
	WalletService   ServiceConfig `mapstructure:"wallet_service"`
	OrderService    ServiceConfig `mapstructure:"order_service"`
	MatchingService ServiceConfig `mapstructure:"matching_service"`
	MarketService   ServiceConfig `mapstructure:"market_service"`
}

type DatabaseConfig struct {
	URI string `mapstructure:"uri"`
}

type RedisConfig struct {
	URI string `mapstructure:"uri"`
}

type NATSConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type OTelConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Insecure bool   `mapstructure:"insecure"`
}

type ServiceConfig struct {
	Address string `mapstructure:"address"`
}

// LoadConfig reads config from YAML file and environment variables.
// Env vars use __ as separator (e.g. DATABASE__URI overrides database.uri).
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()
	v.SetDefault("cors__origins", "*")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Bind known env vars explicitly for nested keys
	envBindings := map[string]string{
		"database.uri":             "DATABASE__URI",
		"redis.uri":                "REDIS__URI",
		"nats.url":                 "NATS__URL",
		"jwt.secret":               "JWT__SECRET",
		"otel.endpoint":            "OTEL__ENDPOINT",
		"wallet_service.address":   "WALLET_SERVICE__ADDRESS",
		"matching_service.address": "MATCHING_SERVICE__ADDRESS",
		"order_service.address":    "ORDER_SERVICE__ADDRESS",
		"cors__origins":            "CORS__ORIGINS",
	}
	for key, env := range envBindings {
		_ = v.BindEnv(key, env)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Defaults
	if cfg.JWT.AccessTTL == 0 {
		cfg.JWT.AccessTTL = 15 * time.Minute
	}
	if cfg.JWT.RefreshTTL == 0 {
		cfg.JWT.RefreshTTL = 7 * 24 * time.Hour
	}
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "text"
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that required configuration fields are set.
func (c *Config) Validate() error {
	if c.Database.URI == "" {
		return fmt.Errorf("DATABASE__URI is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT__SECRET is required")
	}
	if c.Redis.URI == "" {
		return fmt.Errorf("REDIS__URI is required")
	}
	return nil
}
