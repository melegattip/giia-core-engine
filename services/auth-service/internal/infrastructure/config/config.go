package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	Security SecurityConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	PoolMode string
	DSN      string
}

type JWTConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type SecurityConfig struct {
	PasswordMinLength int
	MaxLoginAttempts  int
	LockoutDuration   time.Duration
	RateLimitPerMin   int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", "localhost"),
			Port:         getEnv("PORT", "8083"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  time.Duration(getEnvAsInt("READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("WRITE_TIMEOUT", 10)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5434"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "postgres"),   // Cambiar a 'postgres' para Supabase
			SSLMode:  getEnv("DB_SSLMODE", "require"), // Cambiar a 'require' para Supabase
			PoolMode: getEnv("DB_POOL_MODE", ""),
		},
		JWT: JWTConfig{
			SecretKey:     getEnv("JWT_SECRET", "financial_resume_secret_key_2024"),
			AccessExpiry:  time.Duration(getEnvAsInt("JWT_ACCESS_EXPIRY_HOURS", 24)) * time.Hour,
			RefreshExpiry: time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 7)) * 24 * time.Hour,
			Issuer:        getEnv("JWT_ISSUER", "users-service"),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@example.com"),
			FromName:     getEnv("FROM_NAME", "Financial Resume"),
		},
		Security: SecurityConfig{
			PasswordMinLength: getEnvAsInt("PASSWORD_MIN_LENGTH", 8),
			MaxLoginAttempts:  getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:   time.Duration(getEnvAsInt("LOCKOUT_DURATION_MINUTES", 15)) * time.Minute,
			RateLimitPerMin:   getEnvAsInt("RATE_LIMIT_PER_MIN", 60),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 1),
		},
	}
}

func (c *Config) GetDatabaseDSN() string {
	if c.Database.DSN != "" {
		return c.Database.DSN
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)

	return dsn
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s, using default %d", key, defaultValue)
	}
	return defaultValue
}
