//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConnection_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	t.Run("should connect successfully with valid DSN", func(t *testing.T) {
		db, err := ConnectWithDSN(ctx, dsn)
		require.NoError(t, err)
		require.NotNil(t, db)

		sqlDB, _ := db.DB()
		err = sqlDB.Ping()
		assert.NoError(t, err)

		sqlDB.Close()
	})

	t.Run("should fail with invalid DSN", func(t *testing.T) {
		invalidDSN := "host=invalid port=9999 user=invalid password=invalid dbname=invalid sslmode=disable"
		db, err := ConnectWithDSN(ctx, invalidDSN)
		assert.Error(t, err)
		assert.Nil(t, db)
	})

	t.Run("should connect with Config struct", func(t *testing.T) {
		config := &Config{
			Host:         "localhost",
			Port:         5433,
			User:         "test_user",
			Password:     "test_pass",
			DatabaseName: "test_db",
			SSLMode:      "disable",
		}

		gormDB := New()
		db, err := gormDB.Connect(ctx, config)
		require.NoError(t, err)
		require.NotNil(t, db)

		sqlDB, _ := db.DB()
		defer sqlDB.Close()

		err = sqlDB.Ping()
		assert.NoError(t, err)
	})
}

func TestDatabaseConnectionPooling_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := ConnectWithDSN(ctx, dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	t.Run("should handle concurrent connections", func(t *testing.T) {
		const numGoroutines = 20

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				err := sqlDB.Ping()
				assert.NoError(t, err)
				done <- true
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		stats := sqlDB.Stats()
		assert.LessOrEqual(t, stats.OpenConnections, 10, "Should not exceed max open connections")
	})
}

func TestDatabaseTransactions_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := ConnectWithDSN(ctx, dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	testutil.CreateTestTable(t, db, "test_transactions")
	defer testutil.DropTestTable(t, db, "test_transactions")

	t.Run("should commit transaction successfully", func(t *testing.T) {
		testutil.DropTestTable(t, db, "test_transactions")
		testutil.CreateTestTable(t, db, "test_transactions")

		tx := db.Begin()
		require.NoError(t, tx.Error)

		tx.Exec("INSERT INTO test_transactions (name) VALUES (?)", "test1")
		tx.Exec("INSERT INTO test_transactions (name) VALUES (?)", "test2")

		err := tx.Commit().Error
		assert.NoError(t, err)

		count := testutil.CountRecords(t, db, "test_transactions")
		assert.Equal(t, int64(2), count)
	})

	t.Run("should rollback transaction on error", func(t *testing.T) {
		testutil.DropTestTable(t, db, "test_transactions")
		testutil.CreateTestTable(t, db, "test_transactions")

		tx := db.Begin()
		require.NoError(t, tx.Error)

		tx.Exec("INSERT INTO test_transactions (name) VALUES (?)", "test3")

		tx.Rollback()

		count := testutil.CountRecords(t, db, "test_transactions")
		assert.Equal(t, int64(0), count)
	})

	t.Run("should handle nested transactions", func(t *testing.T) {
		testutil.DropTestTable(t, db, "test_transactions")
		testutil.CreateTestTable(t, db, "test_transactions")

		tx := db.Begin()
		require.NoError(t, tx.Error)

		tx.Exec("INSERT INTO test_transactions (name) VALUES (?)", "outer")

		tx.SavePoint("sp1")
		tx.Exec("INSERT INTO test_transactions (name) VALUES (?)", "inner")

		tx.RollbackTo("sp1")

		err := tx.Commit().Error
		assert.NoError(t, err)

		count := testutil.CountRecords(t, db, "test_transactions")
		assert.Equal(t, int64(1), count)
	})
}

func TestDatabaseHealthCheck_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := ConnectWithDSN(ctx, dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	gormDB := New()

	t.Run("should return healthy when database is up", func(t *testing.T) {
		err := gormDB.HealthCheck(ctx, db)
		assert.NoError(t, err)
	})

	t.Run("should return error when database is closed", func(t *testing.T) {
		testDB, err := ConnectWithDSN(ctx, dsn)
		require.NoError(t, err)

		testSqlDB, _ := testDB.DB()
		testSqlDB.Close()

		err = gormDB.HealthCheck(ctx, testDB)
		assert.Error(t, err)
	})
}

func TestDatabaseRetryLogic_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := ConnectWithDSN(ctx, dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	t.Run("should handle context timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := db.WithContext(timeoutCtx).Raw("SELECT pg_sleep(1)").Error

		assert.Error(t, err)
	})

	t.Run("should successfully execute query within timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var result int
		err := db.WithContext(timeoutCtx).Raw("SELECT 1").Scan(&result).Error

		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})
}

func TestDatabaseClose_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	t.Run("should close database connection successfully", func(t *testing.T) {
		db, err := ConnectWithDSN(ctx, dsn)
		require.NoError(t, err)

		gormDB := New()
		err = gormDB.Close(db)
		assert.NoError(t, err)

		sqlDB, _ := db.DB()
		err = sqlDB.Ping()
		assert.Error(t, err)
	})
}
