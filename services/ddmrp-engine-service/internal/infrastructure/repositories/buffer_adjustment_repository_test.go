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

func TestBufferAdjustmentRepository_Create_ValidationError(t *testing.T) {
	db, _ := setupTestDB(t)

	adjustment := &domain.BufferAdjustment{
		BufferID: uuid.Nil,
	}

	repo := NewBufferAdjustmentRepository(db)
	err := repo.Create(context.Background(), adjustment)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buffer_id is required")
}

func TestBufferAdjustmentRepository_Create_DatabaseError(t *testing.T) {
	db, mock := setupTestDB(t)

	adjustment := &domain.BufferAdjustment{
		ID:             uuid.New(),
		BufferID:       uuid.New(),
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		AdjustmentType: domain.BufferAdjustmentZoneFactor,
		TargetZone:     domain.ZoneGreen,
		Factor:         1.2,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		Reason:         "Test",
		CreatedBy:      uuid.New(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "buffer_adjustments"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	repo := NewBufferAdjustmentRepository(db)
	err := repo.Create(context.Background(), adjustment)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferAdjustmentRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)

	adjustmentID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "buffer_adjustments"`).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewBufferAdjustmentRepository(db)
	result, err := repo.GetByID(context.Background(), adjustmentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferAdjustmentRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	adjustmentID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "buffer_adjustments"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewBufferAdjustmentRepository(db)
	err := repo.Delete(context.Background(), adjustmentID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
