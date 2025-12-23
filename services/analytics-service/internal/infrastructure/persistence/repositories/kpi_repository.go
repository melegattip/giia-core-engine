package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
)

type PostgresKPIRepository struct {
	db *sql.DB
}

func NewPostgresKPIRepository(db *sql.DB) *PostgresKPIRepository {
	return &PostgresKPIRepository{db: db}
}

func (r *PostgresKPIRepository) SaveDaysInInventoryKPI(ctx context.Context, kpi *domain.DaysInInventoryKPI) error {
	query := `
		INSERT INTO days_in_inventory_kpi (
			id, organization_id, snapshot_date, total_valued_days,
			average_valued_days, total_products, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (organization_id, snapshot_date)
		DO UPDATE SET
			total_valued_days = EXCLUDED.total_valued_days,
			average_valued_days = EXCLUDED.average_valued_days,
			total_products = EXCLUDED.total_products,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		kpi.ID, kpi.OrganizationID, kpi.SnapshotDate,
		kpi.TotalValuedDays, kpi.AverageValuedDays, kpi.TotalProducts,
		kpi.CreatedAt, time.Now(),
	)

	return err
}

func (r *PostgresKPIRepository) GetDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.DaysInInventoryKPI, error) {
	query := `
		SELECT id, organization_id, snapshot_date, total_valued_days,
			average_valued_days, total_products, created_at
		FROM days_in_inventory_kpi
		WHERE organization_id = $1 AND snapshot_date = $2
	`

	kpi := &domain.DaysInInventoryKPI{}
	err := r.db.QueryRowContext(ctx, query, organizationID, date).Scan(
		&kpi.ID, &kpi.OrganizationID, &kpi.SnapshotDate,
		&kpi.TotalValuedDays, &kpi.AverageValuedDays, &kpi.TotalProducts,
		&kpi.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewNotFoundError("days_in_inventory_kpi")
	}

	if err != nil {
		return nil, err
	}

	return kpi, nil
}

func (r *PostgresKPIRepository) ListDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.DaysInInventoryKPI, error) {
	query := `
		SELECT id, organization_id, snapshot_date, total_valued_days,
			average_valued_days, total_products, created_at
		FROM days_in_inventory_kpi
		WHERE organization_id = $1 AND snapshot_date BETWEEN $2 AND $3
		ORDER BY snapshot_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kpis []*domain.DaysInInventoryKPI
	for rows.Next() {
		kpi := &domain.DaysInInventoryKPI{}
		err := rows.Scan(
			&kpi.ID, &kpi.OrganizationID, &kpi.SnapshotDate,
			&kpi.TotalValuedDays, &kpi.AverageValuedDays, &kpi.TotalProducts,
			&kpi.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		kpis = append(kpis, kpi)
	}

	return kpis, rows.Err()
}

func (r *PostgresKPIRepository) SaveImmobilizedInventoryKPI(ctx context.Context, kpi *domain.ImmobilizedInventoryKPI) error {
	query := `
		INSERT INTO immobilized_inventory_kpi (
			id, organization_id, snapshot_date, threshold_years,
			immobilized_count, immobilized_value, total_stock_value,
			immobilized_percentage, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (organization_id, snapshot_date, threshold_years)
		DO UPDATE SET
			immobilized_count = EXCLUDED.immobilized_count,
			immobilized_value = EXCLUDED.immobilized_value,
			total_stock_value = EXCLUDED.total_stock_value,
			immobilized_percentage = EXCLUDED.immobilized_percentage,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		kpi.ID, kpi.OrganizationID, kpi.SnapshotDate, kpi.ThresholdYears,
		kpi.ImmobilizedCount, kpi.ImmobilizedValue, kpi.TotalStockValue,
		kpi.ImmobilizedPercentage, kpi.CreatedAt, time.Now(),
	)

	return err
}

func (r *PostgresKPIRepository) GetImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time, thresholdYears int) (*domain.ImmobilizedInventoryKPI, error) {
	query := `
		SELECT id, organization_id, snapshot_date, threshold_years,
			immobilized_count, immobilized_value, total_stock_value,
			immobilized_percentage, created_at
		FROM immobilized_inventory_kpi
		WHERE organization_id = $1 AND snapshot_date = $2 AND threshold_years = $3
	`

	kpi := &domain.ImmobilizedInventoryKPI{}
	err := r.db.QueryRowContext(ctx, query, organizationID, date, thresholdYears).Scan(
		&kpi.ID, &kpi.OrganizationID, &kpi.SnapshotDate, &kpi.ThresholdYears,
		&kpi.ImmobilizedCount, &kpi.ImmobilizedValue, &kpi.TotalStockValue,
		&kpi.ImmobilizedPercentage, &kpi.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewNotFoundError("immobilized_inventory_kpi")
	}

	if err != nil {
		return nil, err
	}

	return kpi, nil
}

func (r *PostgresKPIRepository) ListImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.ImmobilizedInventoryKPI, error) {
	query := `
		SELECT id, organization_id, snapshot_date, threshold_years,
			immobilized_count, immobilized_value, total_stock_value,
			immobilized_percentage, created_at
		FROM immobilized_inventory_kpi
		WHERE organization_id = $1 AND snapshot_date BETWEEN $2 AND $3
		ORDER BY snapshot_date DESC, threshold_years ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kpis []*domain.ImmobilizedInventoryKPI
	for rows.Next() {
		kpi := &domain.ImmobilizedInventoryKPI{}
		err := rows.Scan(
			&kpi.ID, &kpi.OrganizationID, &kpi.SnapshotDate, &kpi.ThresholdYears,
			&kpi.ImmobilizedCount, &kpi.ImmobilizedValue, &kpi.TotalStockValue,
			&kpi.ImmobilizedPercentage, &kpi.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		kpis = append(kpis, kpi)
	}

	return kpis, rows.Err()
}

func (r *PostgresKPIRepository) SaveInventoryRotationKPI(ctx context.Context, kpi *domain.InventoryRotationKPI) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO inventory_rotation_kpi (
			id, organization_id, snapshot_date, sales_last_30_days,
			avg_monthly_stock, rotation_ratio, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (organization_id, snapshot_date)
		DO UPDATE SET
			sales_last_30_days = EXCLUDED.sales_last_30_days,
			avg_monthly_stock = EXCLUDED.avg_monthly_stock,
			rotation_ratio = EXCLUDED.rotation_ratio,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	var kpiID uuid.UUID
	err = tx.QueryRowContext(ctx, query,
		kpi.ID, kpi.OrganizationID, kpi.SnapshotDate,
		kpi.SalesLast30Days, kpi.AvgMonthlyStock, kpi.RotationRatio,
		kpi.CreatedAt, time.Now(),
	).Scan(&kpiID)

	if err != nil {
		return err
	}

	deleteQuery := `DELETE FROM rotating_products WHERE kpi_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, kpiID)
	if err != nil {
		return err
	}

	if len(kpi.TopRotatingProducts) > 0 || len(kpi.SlowRotatingProducts) > 0 {
		insertProductQuery := `
			INSERT INTO rotating_products (
				id, kpi_id, product_id, sku, name, sales_30_days,
				avg_stock_value, rotation_ratio, category, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`

		for _, product := range kpi.TopRotatingProducts {
			_, err = tx.ExecContext(ctx, insertProductQuery,
				uuid.New(), kpiID, product.ProductID, product.SKU, product.Name,
				product.Sales30Days, product.AvgStockValue, product.RotationRatio,
				"top", time.Now(),
			)
			if err != nil {
				return err
			}
		}

		for _, product := range kpi.SlowRotatingProducts {
			_, err = tx.ExecContext(ctx, insertProductQuery,
				uuid.New(), kpiID, product.ProductID, product.SKU, product.Name,
				product.Sales30Days, product.AvgStockValue, product.RotationRatio,
				"slow", time.Now(),
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PostgresKPIRepository) GetInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.InventoryRotationKPI, error) {
	query := `
		SELECT id, organization_id, snapshot_date, sales_last_30_days,
			avg_monthly_stock, rotation_ratio, created_at
		FROM inventory_rotation_kpi
		WHERE organization_id = $1 AND snapshot_date = $2
	`

	kpi := &domain.InventoryRotationKPI{}
	err := r.db.QueryRowContext(ctx, query, organizationID, date).Scan(
		&kpi.ID, &kpi.OrganizationID, &kpi.SnapshotDate,
		&kpi.SalesLast30Days, &kpi.AvgMonthlyStock, &kpi.RotationRatio,
		&kpi.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewNotFoundError("inventory_rotation_kpi")
	}

	if err != nil {
		return nil, err
	}

	productQuery := `
		SELECT product_id, sku, name, sales_30_days, avg_stock_value,
			rotation_ratio, category
		FROM rotating_products
		WHERE kpi_id = $1
		ORDER BY rotation_ratio DESC
	`

	rows, err := r.db.QueryContext(ctx, productQuery, kpi.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topProducts []domain.RotatingProduct
	var slowProducts []domain.RotatingProduct

	for rows.Next() {
		var product domain.RotatingProduct
		var category string
		err := rows.Scan(
			&product.ProductID, &product.SKU, &product.Name,
			&product.Sales30Days, &product.AvgStockValue, &product.RotationRatio,
			&category,
		)
		if err != nil {
			return nil, err
		}

		if category == "top" {
			topProducts = append(topProducts, product)
		} else {
			slowProducts = append(slowProducts, product)
		}
	}

	kpi.TopRotatingProducts = topProducts
	kpi.SlowRotatingProducts = slowProducts

	return kpi, rows.Err()
}

func (r *PostgresKPIRepository) ListInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.InventoryRotationKPI, error) {
	query := `
		SELECT id, organization_id, snapshot_date, sales_last_30_days,
			avg_monthly_stock, rotation_ratio, created_at
		FROM inventory_rotation_kpi
		WHERE organization_id = $1 AND snapshot_date BETWEEN $2 AND $3
		ORDER BY snapshot_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kpis []*domain.InventoryRotationKPI
	for rows.Next() {
		kpi := &domain.InventoryRotationKPI{}
		err := rows.Scan(
			&kpi.ID, &kpi.OrganizationID, &kpi.SnapshotDate,
			&kpi.SalesLast30Days, &kpi.AvgMonthlyStock, &kpi.RotationRatio,
			&kpi.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		kpis = append(kpis, kpi)
	}

	return kpis, rows.Err()
}

func (r *PostgresKPIRepository) SaveBufferAnalytics(ctx context.Context, analytics *domain.BufferAnalytics) error {
	query := `
		INSERT INTO buffer_analytics (
			id, product_id, organization_id, snapshot_date, cpd,
			red_zone, red_base, red_safe, yellow_zone, green_zone,
			ltd, lead_time_factor, variability_factor, moq, order_frequency,
			optimal_order_freq, safety_days, avg_open_orders, has_adjustments,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (product_id, organization_id, snapshot_date)
		DO UPDATE SET
			cpd = EXCLUDED.cpd,
			red_zone = EXCLUDED.red_zone,
			red_base = EXCLUDED.red_base,
			red_safe = EXCLUDED.red_safe,
			yellow_zone = EXCLUDED.yellow_zone,
			green_zone = EXCLUDED.green_zone,
			ltd = EXCLUDED.ltd,
			lead_time_factor = EXCLUDED.lead_time_factor,
			variability_factor = EXCLUDED.variability_factor,
			moq = EXCLUDED.moq,
			order_frequency = EXCLUDED.order_frequency,
			optimal_order_freq = EXCLUDED.optimal_order_freq,
			safety_days = EXCLUDED.safety_days,
			avg_open_orders = EXCLUDED.avg_open_orders,
			has_adjustments = EXCLUDED.has_adjustments,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		analytics.ID, analytics.ProductID, analytics.OrganizationID, analytics.Date,
		analytics.CPD, analytics.RedZone, analytics.RedBase, analytics.RedSafe,
		analytics.YellowZone, analytics.GreenZone, analytics.LTD,
		analytics.LeadTimeFactor, analytics.VariabilityFactor,
		analytics.MOQ, analytics.OrderFrequency,
		analytics.OptimalOrderFreq, analytics.SafetyDays, analytics.AvgOpenOrders,
		analytics.HasAdjustments, analytics.CreatedAt, time.Now(),
	)

	return err
}

func (r *PostgresKPIRepository) GetBufferAnalytics(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) (*domain.BufferAnalytics, error) {
	query := `
		SELECT id, product_id, organization_id, snapshot_date, cpd,
			red_zone, red_base, red_safe, yellow_zone, green_zone,
			ltd, lead_time_factor, variability_factor, moq, order_frequency,
			optimal_order_freq, safety_days, avg_open_orders, has_adjustments, created_at
		FROM buffer_analytics
		WHERE product_id = $1 AND organization_id = $2 AND snapshot_date = $3
	`

	analytics := &domain.BufferAnalytics{}
	err := r.db.QueryRowContext(ctx, query, productID, organizationID, date).Scan(
		&analytics.ID, &analytics.ProductID, &analytics.OrganizationID, &analytics.Date,
		&analytics.CPD, &analytics.RedZone, &analytics.RedBase, &analytics.RedSafe,
		&analytics.YellowZone, &analytics.GreenZone, &analytics.LTD,
		&analytics.LeadTimeFactor, &analytics.VariabilityFactor,
		&analytics.MOQ, &analytics.OrderFrequency,
		&analytics.OptimalOrderFreq, &analytics.SafetyDays, &analytics.AvgOpenOrders,
		&analytics.HasAdjustments, &analytics.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewNotFoundError("buffer_analytics")
	}

	if err != nil {
		return nil, err
	}

	return analytics, nil
}

func (r *PostgresKPIRepository) ListBufferAnalytics(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.BufferAnalytics, error) {
	query := `
		SELECT id, product_id, organization_id, snapshot_date, cpd,
			red_zone, red_base, red_safe, yellow_zone, green_zone,
			ltd, lead_time_factor, variability_factor, moq, order_frequency,
			optimal_order_freq, safety_days, avg_open_orders, has_adjustments, created_at
		FROM buffer_analytics
		WHERE organization_id = $1 AND snapshot_date BETWEEN $2 AND $3
		ORDER BY snapshot_date DESC, product_id
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyticsList []*domain.BufferAnalytics
	for rows.Next() {
		analytics := &domain.BufferAnalytics{}
		err := rows.Scan(
			&analytics.ID, &analytics.ProductID, &analytics.OrganizationID, &analytics.Date,
			&analytics.CPD, &analytics.RedZone, &analytics.RedBase, &analytics.RedSafe,
			&analytics.YellowZone, &analytics.GreenZone, &analytics.LTD,
			&analytics.LeadTimeFactor, &analytics.VariabilityFactor,
			&analytics.MOQ, &analytics.OrderFrequency,
			&analytics.OptimalOrderFreq, &analytics.SafetyDays, &analytics.AvgOpenOrders,
			&analytics.HasAdjustments, &analytics.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		analyticsList = append(analyticsList, analytics)
	}

	return analyticsList, rows.Err()
}

func (r *PostgresKPIRepository) SaveKPISnapshot(ctx context.Context, snapshot *domain.KPISnapshot) error {
	query := `
		INSERT INTO kpi_snapshots (
			id, organization_id, snapshot_date, inventory_turnover,
			stockout_rate, service_level, excess_inventory_pct,
			buffer_score_green, buffer_score_yellow, buffer_score_red,
			total_inventory_value, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (organization_id, snapshot_date)
		DO UPDATE SET
			inventory_turnover = EXCLUDED.inventory_turnover,
			stockout_rate = EXCLUDED.stockout_rate,
			service_level = EXCLUDED.service_level,
			excess_inventory_pct = EXCLUDED.excess_inventory_pct,
			buffer_score_green = EXCLUDED.buffer_score_green,
			buffer_score_yellow = EXCLUDED.buffer_score_yellow,
			buffer_score_red = EXCLUDED.buffer_score_red,
			total_inventory_value = EXCLUDED.total_inventory_value,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		snapshot.ID, snapshot.OrganizationID, snapshot.SnapshotDate,
		snapshot.InventoryTurnover, snapshot.StockoutRate, snapshot.ServiceLevel,
		snapshot.ExcessInventoryPct, snapshot.BufferScoreGreen,
		snapshot.BufferScoreYellow, snapshot.BufferScoreRed,
		snapshot.TotalInventoryValue, snapshot.CreatedAt, time.Now(),
	)

	return err
}

func (r *PostgresKPIRepository) GetKPISnapshot(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.KPISnapshot, error) {
	query := `
		SELECT id, organization_id, snapshot_date, inventory_turnover,
			stockout_rate, service_level, excess_inventory_pct,
			buffer_score_green, buffer_score_yellow, buffer_score_red,
			total_inventory_value, created_at
		FROM kpi_snapshots
		WHERE organization_id = $1 AND snapshot_date = $2
	`

	snapshot := &domain.KPISnapshot{}
	err := r.db.QueryRowContext(ctx, query, organizationID, date).Scan(
		&snapshot.ID, &snapshot.OrganizationID, &snapshot.SnapshotDate,
		&snapshot.InventoryTurnover, &snapshot.StockoutRate, &snapshot.ServiceLevel,
		&snapshot.ExcessInventoryPct, &snapshot.BufferScoreGreen,
		&snapshot.BufferScoreYellow, &snapshot.BufferScoreRed,
		&snapshot.TotalInventoryValue, &snapshot.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewNotFoundError("kpi_snapshot")
	}

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (r *PostgresKPIRepository) ListKPISnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.KPISnapshot, error) {
	query := `
		SELECT id, organization_id, snapshot_date, inventory_turnover,
			stockout_rate, service_level, excess_inventory_pct,
			buffer_score_green, buffer_score_yellow, buffer_score_red,
			total_inventory_value, created_at
		FROM kpi_snapshots
		WHERE organization_id = $1 AND snapshot_date BETWEEN $2 AND $3
		ORDER BY snapshot_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*domain.KPISnapshot
	for rows.Next() {
		snapshot := &domain.KPISnapshot{}
		err := rows.Scan(
			&snapshot.ID, &snapshot.OrganizationID, &snapshot.SnapshotDate,
			&snapshot.InventoryTurnover, &snapshot.StockoutRate, &snapshot.ServiceLevel,
			&snapshot.ExcessInventoryPct, &snapshot.BufferScoreGreen,
			&snapshot.BufferScoreYellow, &snapshot.BufferScoreRed,
			&snapshot.TotalInventoryValue, &snapshot.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}
