package repositories

import (
	"context"
	"errors"
	"regexp"
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

func setupADUTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func TestADURepository_Create_Success(t *testing.T) {
	db, mock := setupADUTestDB(t)

	aduID := uuid.New()
	productID := uuid.New()
	orgID := uuid.New()
	now := time.Now()

	adu := &domain.ADUCalculation{
		ID:              aduID,
		ProductID:       productID,
		OrganizationID:  orgID,
		CalculationDate: now,
		ADUValue:        100.0,
		Method:          domain.ADUMethodAverage,
		PeriodDays:      30,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "adu_calculations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(aduID, now))
	mock.ExpectCommit()

	repo := NewADURepository(db)
	err := repo.Create(context.Background(), adu)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestADURepository_Create_ValidationError(t *testing.T) {
	db, _ := setupADUTestDB(t)

	adu := &domain.ADUCalculation{
		ProductID: uuid.Nil,
	}

	repo := NewADURepository(db)
	err := repo.Create(context.Background(), adu)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestADURepository_Create_DatabaseError(t *testing.T) {
	db, mock := setupADUTestDB(t)

	adu := &domain.ADUCalculation{
		ID:              uuid.New(),
		ProductID:       uuid.New(),
		OrganizationID:  uuid.New(),
		CalculationDate: time.Now(),
		ADUValue:        100.0,
		Method:          domain.ADUMethodAverage,
		PeriodDays:      30,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "adu_calculations"`)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	repo := NewADURepository(db)
	err := repo.Create(context.Background(), adu)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestADURepository_GetLatest_NotFound(t *testing.T) {
	db, mock := setupADUTestDB(t)

	productID := uuid.New()
	orgID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "adu_calculations"`).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewADURepository(db)
	result, err := repo.GetLatest(context.Background(), productID, orgID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestADURepository_Delete_Success(t *testing.T) {
	db, mock := setupADUTestDB(t)

	aduID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "adu_calculations"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewADURepository(db)
	err := repo.Delete(context.Background(), aduID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
