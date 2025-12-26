// Package integration provides integration testing utilities for the GIIA platform.
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Teardown cleans up the test environment.
func (env *TestEnvironment) Teardown() error {
	fmt.Println("🧹 Tearing down integration test environment...")

	// Cancel the context
	if env.Cancel != nil {
		env.Cancel()
	}

	fmt.Println("✅ Teardown complete")
	return nil
}

// CleanupDatabase removes all test data from the database.
func (env *TestEnvironment) CleanupDatabase() error {
	fmt.Println("🗑️  Cleaning up test database...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", env.PostgresURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Tables to clean in order (respecting foreign keys)
	tables := []string{
		// AI Hub tables
		"ai_hub_notifications",
		"ai_hub_user_preferences",
		"ai_hub_optimization_hints",
		"ai_hub_events",

		// Analytics tables
		"analytics_snapshots",
		"analytics_kpi_history",

		// Execution tables
		"execution_inventory_transactions",
		"execution_inventory_balances",
		"execution_sales_order_items",
		"execution_sales_orders",
		"execution_purchase_order_items",
		"execution_purchase_orders",
		"execution_alerts",

		// DDMRP tables
		"ddmrp_demand_adjustments",
		"ddmrp_buffers",
		"ddmrp_buffer_profiles",

		// Catalog tables
		"catalog_product_attributes",
		"catalog_products",
		"catalog_categories",
		"catalog_suppliers",

		// Auth tables
		"auth_role_permissions",
		"auth_user_roles",
		"auth_permissions",
		"auth_roles",
		"auth_sessions",
		"auth_users",
		"auth_organizations",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Table might not exist, continue
			fmt.Printf("  ⚠️  Could not truncate %s: %v\n", table, err)
		}
	}

	fmt.Println("✅ Database cleanup complete")
	return nil
}

// CleanupTestData cleans up data created during a specific test.
func (env *TestEnvironment) CleanupTestData(organizationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", env.PostgresURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Delete data for specific organization
	deleteQueries := []string{
		"DELETE FROM ai_hub_notifications WHERE organization_id = $1",
		"DELETE FROM ai_hub_user_preferences WHERE organization_id = $1",
		"DELETE FROM analytics_snapshots WHERE organization_id = $1",
		"DELETE FROM execution_inventory_transactions WHERE organization_id = $1",
		"DELETE FROM execution_inventory_balances WHERE organization_id = $1",
		"DELETE FROM execution_sales_order_items WHERE organization_id = $1",
		"DELETE FROM execution_sales_orders WHERE organization_id = $1",
		"DELETE FROM execution_purchase_order_items WHERE organization_id = $1",
		"DELETE FROM execution_purchase_orders WHERE organization_id = $1",
		"DELETE FROM ddmrp_demand_adjustments WHERE organization_id = $1",
		"DELETE FROM ddmrp_buffers WHERE organization_id = $1",
		"DELETE FROM catalog_products WHERE organization_id = $1",
		"DELETE FROM auth_users WHERE organization_id = $1",
	}

	for _, query := range deleteQueries {
		_, err := db.ExecContext(ctx, query, organizationID)
		if err != nil {
			// Continue on error (table might not exist or no data)
			continue
		}
	}

	return nil
}

// ResetSequences resets all database sequences (useful for deterministic testing).
func (env *TestEnvironment) ResetSequences() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", env.PostgresURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Get all sequences and reset them
	query := `
		SELECT sequencename FROM pg_sequences WHERE schemaname = 'public'
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get sequences: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var seqName string
		if err := rows.Scan(&seqName); err != nil {
			continue
		}

		_, err := db.ExecContext(ctx, fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seqName))
		if err != nil {
			fmt.Printf("  ⚠️  Could not reset sequence %s: %v\n", seqName, err)
		}
	}

	return nil
}
