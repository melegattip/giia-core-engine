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

func TestBufferHistoryRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	historyID := uuid.New()
	now := time.Now()

	history := &domain.BufferHistory{
		ID:             historyID,
		BufferID:       uuid.New(),
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		SnapshotDate:   now,
		CPD:            100.0,
		DLT:            30,
		RedZone:        2250.0,
		YellowZone:     3000.0,
		GreenZone:      2000.0,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "buffer_history"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(historyID, now))
	mock.ExpectCommit()

	repo := NewBufferHistoryRepository(db)
	err := repo.Create(context.Background(), history)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferHistoryRepository_Create_DatabaseError(t *testing.T) {
	db, mock := setupTestDB(t)

	history := &domain.BufferHistory{
		ID:           uuid.New(),
		BufferID:     uuid.New(),
		SnapshotDate: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "buffer_history"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	repo := NewBufferHistoryRepository(db)
	err := repo.Create(context.Background(), history)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferHistoryRepository_GetByBufferAndDate_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)

	bufferID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "buffer_history"`).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewBufferHistoryRepository(db)
	result, err := repo.GetByBufferAndDate(context.Background(), bufferID, time.Now())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBufferHistoryRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)

	historyID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "buffer_history"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewBufferHistoryRepository(db)
	err := repo.Delete(context.Background(), historyID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
