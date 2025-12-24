package buffer

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBuffer_Execute_Success(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	productID := uuid.New()
	orgID := uuid.New()
	expectedBuffer := &domain.Buffer{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		CPD:            100.0,
		LTD:            30,
		Zone:           domain.ZoneGreen,
	}

	mockBufferRepo.On("GetByProduct", mock.Anything, productID, orgID).
		Return(expectedBuffer, nil)

	useCase := NewGetBufferUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), productID, orgID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedBuffer.ID, result.ID)
	assert.Equal(t, expectedBuffer.ProductID, result.ProductID)
	assert.Equal(t, expectedBuffer.CPD, result.CPD)

	mockBufferRepo.AssertExpectations(t)
}

func TestGetBuffer_Execute_BufferNotFound(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	productID := uuid.New()
	orgID := uuid.New()

	mockBufferRepo.On("GetByProduct", mock.Anything, productID, orgID).
		Return(nil, domain.NewNotFoundError("buffer not found"))

	useCase := NewGetBufferUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), productID, orgID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "buffer not found")

	mockBufferRepo.AssertExpectations(t)
}

func TestGetBuffer_Execute_NilProductID(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	useCase := NewGetBufferUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), uuid.Nil, uuid.New())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestGetBuffer_Execute_NilOrganizationID(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	useCase := NewGetBufferUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestGetBuffer_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		productID      uuid.UUID
		organizationID uuid.UUID
		setupMock      func(m *MockBufferRepository)
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success",
			productID:      uuid.New(),
			organizationID: uuid.New(),
			setupMock: func(m *MockBufferRepository) {
				m.On("GetByProduct", mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.Buffer{ID: uuid.New(), CPD: 100.0, LTD: 30}, nil)
			},
			wantErr: false,
		},
		{
			name:           "nil product_id",
			productID:      uuid.Nil,
			organizationID: uuid.New(),
			setupMock:      func(m *MockBufferRepository) {},
			wantErr:        true,
			errContains:    "product_id is required",
		},
		{
			name:           "nil organization_id",
			productID:      uuid.New(),
			organizationID: uuid.Nil,
			setupMock:      func(m *MockBufferRepository) {},
			wantErr:        true,
			errContains:    "organization_id is required",
		},
		{
			name:           "buffer not found",
			productID:      uuid.New(),
			organizationID: uuid.New(),
			setupMock: func(m *MockBufferRepository) {
				m.On("GetByProduct", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, domain.NewNotFoundError("buffer not found"))
			},
			wantErr:     true,
			errContains: "buffer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBufferRepo := new(MockBufferRepository)
			tt.setupMock(mockBufferRepo)

			useCase := NewGetBufferUseCase(mockBufferRepo)
			result, err := useCase.Execute(context.Background(), tt.productID, tt.organizationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockBufferRepo.AssertExpectations(t)
		})
	}
}
