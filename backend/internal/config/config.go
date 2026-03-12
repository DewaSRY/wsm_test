package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Marketplace MarketplaceConfig
	MarketPlaceAuth MarketPlaceAuth
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	Schema          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type MarketplaceConfig struct {
	BaseURL       string
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
}

type WorkerConfig struct {
	OrderSyncInterval   time.Duration
	TokenRefreshBuffer  time.Duration
	WebhookRetryLimit   int
	WebhookRetryDelay   time.Duration
}

type MarketPlaceAuth struct {
	ShopId string
	State string
	PartnerId string
	Timestamp string
	Sign string
	Redirect string
 }

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnvString("DB_HOST", "localhost"),
			Port:            getEnvString("DB_PORT", "5432"),
			Username:        getEnvString("DB_USERNAME", "postgres"),
			Password:        getEnvString("DB_PASSWORD", ""),
			Database:        getEnvString("DB_DATABASE", "wms"),
			Schema:          getEnvString("DB_SCHEMA", "public"),
			SSLMode:         getEnvString("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Marketplace: MarketplaceConfig{
			BaseURL:       getEnvString("MARKETPLACE_BASE_URL", "https://fullstack-technical-test.suksescorp.co.id"),
			Timeout:       getEnvDuration("MARKETPLACE_TIMEOUT", 30*time.Second),
			RetryAttempts: getEnvInt("MARKETPLACE_RETRY_ATTEMPTS", 3),
			RetryDelay:    getEnvDuration("MARKETPLACE_RETRY_DELAY", 1*time.Second),
		},
		MarketPlaceAuth: MarketPlaceAuth{
			ShopId:  getEnvString("SHOP_ID", "shopee-123"),
			State:  getEnvString("STATE", "test"),
			PartnerId:  getEnvString("PARTNER_ID", "992800"),
			Timestamp:  getEnvString("TIMESTAMP", "1773280806"),
			Sign:  getEnvString("SIGN", "5035ab502857d829449c4dd391c9563c32698789296687d8f8879fa81934d7ff"),
			Redirect:  getEnvString("REDIRECT", "https://example.com/callback"),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "postgres://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Database + "?sslmode=" + c.SSLMode + "&search_path=" + c.Schema
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
