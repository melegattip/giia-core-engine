package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DatabaseName    string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	SlowQueryTime   time.Duration
}

type Database interface {
	Connect(ctx context.Context, config *Config) (*gorm.DB, error)
	HealthCheck(ctx context.Context, db *gorm.DB) error
	Close(db *gorm.DB) error
}

type GormDatabase struct{}

func New() *GormDatabase {
	return &GormDatabase{}
}

func (d *GormDatabase) Connect(ctx context.Context, config *Config) (*gorm.DB, error) {
	dsn := buildDSN(config)

	var db *gorm.DB
	var err error

	err = retryWithBackoff(ctx, func() error {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
		if err != nil {
			return err
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		if config.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		} else {
			sqlDB.SetMaxOpenConns(25)
		}

		if config.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		} else {
			sqlDB.SetMaxIdleConns(5)
		}

		if config.ConnMaxLifetime > 0 {
			sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
		} else {
			sqlDB.SetConnMaxLifetime(5 * time.Minute)
		}

		return sqlDB.PingContext(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
	}

	return db, nil
}

func (d *GormDatabase) HealthCheck(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

func (d *GormDatabase) Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	return sqlDB.Close()
}

func buildDSN(config *Config) string {
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DatabaseName,
		sslMode,
	)
}

func ConnectWithDSN(ctx context.Context, dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	err = retryWithBackoff(ctx, func() error {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
		if err != nil {
			return err
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)

		return sqlDB.PingContext(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
	}

	return db, nil
}
