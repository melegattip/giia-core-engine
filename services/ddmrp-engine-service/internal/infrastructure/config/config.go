package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Catalog  CatalogConfig
	NATS     NATSConfig
	Cron     CronConfig
}

type ServerConfig struct {
	GRPCPort string
	HTTPPort string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type CatalogConfig struct {
	GRPCURL string
}

type NATSConfig struct {
	URL     string
	Enabled bool
}

type CronConfig struct {
	Enabled bool
	Schedule string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			GRPCPort: getEnv("GRPC_PORT", "50053"),
			HTTPPort: getEnv("HTTP_PORT", "8083"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "giia_ddmrp"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Catalog: CatalogConfig{
			GRPCURL: getEnv("CATALOG_GRPC_URL", "localhost:50051"),
		},
		NATS: NATSConfig{
			URL:     getEnv("NATS_URL", "nats://localhost:4222"),
			Enabled: getEnvBool("NATS_ENABLED", false),
		},
		Cron: CronConfig{
			Enabled:  getEnvBool("CRON_ENABLED", true),
			Schedule: getEnv("CRON_SCHEDULE", "0 2 * * *"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func (c *Config) GetDSN() string {
	return "host=" + c.Database.Host +
		" port=" + c.Database.Port +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Name +
		" sslmode=" + c.Database.SSLMode
}
