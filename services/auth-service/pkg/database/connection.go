package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
)

type DB struct {
	*sql.DB
}

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewConnection(config Config) (*DB, error) {
	// Usar configuración estándar para Supabase
	dsn := config.DSN
	log.Println("🔧 [Database] Using standard Supabase configuration")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to open database connection")
	}

	// Configure connection pool
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(25) // Default
	}

	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(5) // Default
	}

	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(5 * time.Minute) // Default
	}

	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	} else {
		db.SetConnMaxIdleTime(2 * time.Minute) // Default
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to ping database")
	}

	log.Println("✅ [Database] Connection established successfully")

	return &DB{db}, nil
}

func (db *DB) Close() error {
	log.Println("🔌 [Database] Closing connection...")
	return db.DB.Close()
}

func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return pkgErrors.NewInternalServerError("database health check failed")
	}

	return nil
}

func (db *DB) RunMigrations() error {
	log.Println("🔄 [Database] Running migrations...")

	migrations := []string{
		createUsersTable,
		createPreferencesTable,
		createNotificationSettingsTable,
		createTwoFATable,
		createUserTokensTable,
		createLoginAttemptsTable,
		createIndexes,
	}

	for i, migration := range migrations {
		log.Printf("Running migration %d/%d...", i+1, len(migrations))
		if _, err := db.Exec(migration); err != nil {
			return pkgErrors.NewInternalServerError("database migration failed")
		}
	}

	log.Println("✅ [Database] All migrations completed successfully")
	return nil
}

// Migration SQL statements
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    email_verification_token VARCHAR(255),
    email_verification_expires TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createPreferencesTable = `
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(3) DEFAULT 'USD',
    language VARCHAR(5) DEFAULT 'en',
    theme VARCHAR(10) DEFAULT 'light',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    timezone VARCHAR(50) DEFAULT 'UTC',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createNotificationSettingsTable = `
CREATE TABLE IF NOT EXISTS user_notification_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_notifications BOOLEAN DEFAULT true,
    push_notifications BOOLEAN DEFAULT true,
    weekly_reports BOOLEAN DEFAULT true,
    expense_alerts BOOLEAN DEFAULT true,
    budget_alerts BOOLEAN DEFAULT true,
    achievement_notifications BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createTwoFATable = `
CREATE TABLE IF NOT EXISTS user_two_fa (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT false,
    backup_codes TEXT[], -- Array of backup codes
    last_used_code VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createUserTokensTable = `
CREATE TABLE IF NOT EXISTS user_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    token_type VARCHAR(20) NOT NULL, -- 'refresh', 'access', 'email_verification', 'password_reset'
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createLoginAttemptsTable = `
CREATE TABLE IF NOT EXISTS login_attempts (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_email_verification_token ON users(email_verification_token);
CREATE INDEX IF NOT EXISTS idx_users_password_reset_token ON users(password_reset_token);
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_token_hash ON user_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_tokens_expires_at ON user_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_email ON login_attempts(email);
CREATE INDEX IF NOT EXISTS idx_login_attempts_attempted_at ON login_attempts(attempted_at);`
