package testutil

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDatabase(t *testing.T, dsn string) (*gorm.DB, func()) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func TruncateAllTables(t *testing.T, db *gorm.DB) {
	var tables []string
	err := db.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables).Error

	if err != nil {
		t.Logf("Failed to get tables: %v", err)
		return
	}

	for _, table := range tables {
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
}

func CreateTestTable(t *testing.T, db *gorm.DB, tableName string) {
	err := db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, tableName)).Error

	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
}

func DropTestTable(t *testing.T, db *gorm.DB, tableName string) {
	err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName)).Error
	if err != nil {
		t.Logf("Failed to drop test table: %v", err)
	}
}

func CountRecords(t *testing.T, db *gorm.DB, tableName string) int64 {
	var count int64
	err := db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count).Error
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}
	return count
}

func WaitForDatabase(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
