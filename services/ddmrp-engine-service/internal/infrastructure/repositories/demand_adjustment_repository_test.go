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
	"gorm.io/gorm"
)

func TestDemandAdjustmentRepository_Create_ValidationError(t *testing.T) {
	db, _ := setupTestDB(t)

	adjustment := &domain.DemandAdjustment{
		ProductID: uuid.Nil,
	}

	repo := NewDemandAdjustmentRepository(db)
	err := repo.Create(context.Background(), adjustment)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestDemandAdjustmentRepository_Create_DatabaseError(t *testing.T) {
	db, mock := setupTestDB(t)

	adjustment := &domain.DemandAdjustment{
		ID:             uuid.New(),
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test",
		CreatedBy:      uuid.New(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "demand_adjustments"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	repo := NewDemandAdjustmentRepository(db)
	err := repo.Create(context.Background(), adjustment)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDemandAdjustmentRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)

	adjustmentID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "demand_adjustments"`).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewDemandAdjustmentRepository(db)
	result, err := repo.GetByID(context.Background(), adjustmentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDemandAdjustmentRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	adjustmentID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "demand_adjustments"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewDemandAdjustmentRepository(db)
	err := repo.Delete(context.Background(), adjustmentID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
