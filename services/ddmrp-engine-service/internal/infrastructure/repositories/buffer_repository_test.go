package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestBufferRepository_Create_ValidationError(t *testing.T) {
	db, _ := setupTestDB(t)

	buffer := &domain.Buffer{
		ProductID: uuid.Nil,
	}

	repo := NewBufferRepository(db)
	err := repo.Create(context.Background(), buffer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestBufferRepository_Create_DatabaseError(t *testing.T) {
	db, mock := setupTestDB(t)

	buffer := &domain.Buffer{
		ID:              uuid.New(),
		ProductID:       uuid.New(),
		OrganizationID:  uuid.New(),
		BufferProfileID: uuid.New(),
		CPD:             100.0,
		LTD:             30,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "buffers"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	repo := NewBufferRepository(db)
	err := repo.Create(context.Background(), buffer)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)

	bufferID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "buffers"`).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewBufferRepository(db)
	result, err := repo.GetByID(context.Background(), bufferID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_GetByProduct_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)

	productID := uuid.New()
	orgID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "buffers"`).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewBufferRepository(db)
	result, err := repo.GetByProduct(context.Background(), productID, orgID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_List_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	orgID := uuid.New()
	now := time.Now()
	bufferID := uuid.New()
	productID := uuid.New()
	profileID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "product_id", "organization_id", "buffer_profile_id",
		"cpd", "ltd", "red_base", "red_safe", "red_zone",
		"yellow_zone", "green_zone", "top_of_red", "top_of_yellow", "top_of_green",
		"on_hand", "on_order", "qualified_demand", "net_flow_position",
		"buffer_penetration", "zone", "alert_level", "last_recalculated_at",
		"created_at", "updated_at",
	}).AddRow(
		bufferID, productID, orgID, profileID,
		100.0, 30, 1500.0, 750.0, 2250.0,
		3000.0, 2000.0, 2250.0, 5250.0, 7250.0,
		5000.0, 1000.0, 500.0, 5500.0,
		75.86, "green", "normal", now,
		now, now,
	)

	mock.ExpectQuery(`SELECT \* FROM "buffers"`).
		WillReturnRows(rows)

	repo := NewBufferRepository(db)
	result, err := repo.List(context.Background(), orgID, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_ListByZone_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	orgID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "product_id", "organization_id", "buffer_profile_id",
		"cpd", "ltd", "red_base", "red_safe", "red_zone",
		"yellow_zone", "green_zone", "top_of_red", "top_of_yellow", "top_of_green",
		"on_hand", "on_order", "qualified_demand", "net_flow_position",
		"buffer_penetration", "zone", "alert_level", "last_recalculated_at",
		"created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT \* FROM "buffers"`).
		WillReturnRows(rows)

	repo := NewBufferRepository(db)
	result, err := repo.ListByZone(context.Background(), orgID, domain.ZoneRed)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_ListByAlertLevel_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	orgID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "product_id", "organization_id", "buffer_profile_id",
		"cpd", "ltd", "red_base", "red_safe", "red_zone",
		"yellow_zone", "green_zone", "top_of_red", "top_of_yellow", "top_of_green",
		"on_hand", "on_order", "qualified_demand", "net_flow_position",
		"buffer_penetration", "zone", "alert_level", "last_recalculated_at",
		"created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT \* FROM "buffers"`).
		WillReturnRows(rows)

	repo := NewBufferRepository(db)
	result, err := repo.ListByAlertLevel(context.Background(), orgID, domain.AlertCritical)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_UpdateNFP_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	bufferID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "buffers" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewBufferRepository(db)
	err := repo.UpdateNFP(context.Background(), bufferID, 500.0, 100.0, 50.0)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	bufferID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "buffers"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewBufferRepository(db)
	err := repo.Delete(context.Background(), bufferID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
