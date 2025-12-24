package demand_adjustment

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListFADs_ExecuteByProduct_Success(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	productID := uuid.New()
	orgID := uuid.New()
	expectedFADs := []domain.DemandAdjustment{
		{
			ID:             uuid.New(),
			ProductID:      productID,
			OrganizationID: orgID,
			StartDate:      time.Now(),
			EndDate:        time.Now().AddDate(0, 1, 0),
			AdjustmentType: domain.DemandAdjustmentFAD,
			Factor:         1.2,
		},
		{
			ID:             uuid.New(),
			ProductID:      productID,
			OrganizationID: orgID,
			StartDate:      time.Now().AddDate(0, 2, 0),
			EndDate:        time.Now().AddDate(0, 3, 0),
			AdjustmentType: domain.DemandAdjustmentSeasonal,
			Factor:         1.5,
		},
	}

	mockRepo.On("ListByProduct", mock.Anything, productID, orgID).
		Return(expectedFADs, nil)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByProduct(context.Background(), ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, productID, result[0].ProductID)
	assert.Equal(t, productID, result[1].ProductID)

	mockRepo.AssertExpectations(t)
}

func TestListFADs_ExecuteByProduct_EmptyResult(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	productID := uuid.New()
	orgID := uuid.New()

	mockRepo.On("ListByProduct", mock.Anything, productID, orgID).
		Return([]domain.DemandAdjustment{}, nil)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByProduct(context.Background(), ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})

	assert.NoError(t, err)
	assert.Empty(t, result)

	mockRepo.AssertExpectations(t)
}

func TestListFADs_ExecuteByProduct_NilProductID(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByProduct(context.Background(), ListFADsByProductInput{
		ProductID:      uuid.Nil,
		OrganizationID: uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestListFADs_ExecuteByProduct_NilOrganizationID(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByProduct(context.Background(), ListFADsByProductInput{
		ProductID:      uuid.New(),
		OrganizationID: uuid.Nil,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestListFADs_ExecuteByProduct_RepositoryError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	productID := uuid.New()
	orgID := uuid.New()

	mockRepo.On("ListByProduct", mock.Anything, productID, orgID).
		Return([]domain.DemandAdjustment{}, domain.NewInternalError("database error"))

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByProduct(context.Background(), ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list demand adjustments")

	mockRepo.AssertExpectations(t)
}

func TestListFADs_ExecuteByOrganization_Success(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	orgID := uuid.New()
	expectedFADs := []domain.DemandAdjustment{
		{ID: uuid.New(), OrganizationID: orgID, ProductID: uuid.New()},
		{ID: uuid.New(), OrganizationID: orgID, ProductID: uuid.New()},
		{ID: uuid.New(), OrganizationID: orgID, ProductID: uuid.New()},
	}

	mockRepo.On("ListByOrganization", mock.Anything, orgID, 10, 0).
		Return(expectedFADs, nil)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByOrganization(context.Background(), ListFADsByOrganizationInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	mockRepo.AssertExpectations(t)
}

func TestListFADs_ExecuteByOrganization_WithPagination(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	orgID := uuid.New()
	expectedFADs := []domain.DemandAdjustment{
		{ID: uuid.New(), OrganizationID: orgID},
	}

	mockRepo.On("ListByOrganization", mock.Anything, orgID, 5, 10).
		Return(expectedFADs, nil)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByOrganization(context.Background(), ListFADsByOrganizationInput{
		OrganizationID: orgID,
		Limit:          5,
		Offset:         10,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)

	mockRepo.AssertExpectations(t)
}

func TestListFADs_ExecuteByOrganization_NilOrganizationID(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByOrganization(context.Background(), ListFADsByOrganizationInput{
		OrganizationID: uuid.Nil,
		Limit:          10,
		Offset:         0,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestListFADs_ExecuteByOrganization_RepositoryError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)

	orgID := uuid.New()

	mockRepo.On("ListByOrganization", mock.Anything, orgID, 10, 0).
		Return([]domain.DemandAdjustment{}, domain.NewInternalError("database error"))

	useCase := NewListFADsUseCase(mockRepo)
	result, err := useCase.ExecuteByOrganization(context.Background(), ListFADsByOrganizationInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list demand adjustments")

	mockRepo.AssertExpectations(t)
}

func TestListFADs_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		testType    string // "byProduct" or "byOrganization"
		input       interface{}
		setupMock   func(m *MockDemandAdjustmentRepository)
		wantLen     int
		wantErr     bool
		errContains string
	}{
		{
			name:     "byProduct - success",
			testType: "byProduct",
			input: ListFADsByProductInput{
				ProductID:      uuid.New(),
				OrganizationID: uuid.New(),
			},
			setupMock: func(m *MockDemandAdjustmentRepository) {
				m.On("ListByProduct", mock.Anything, mock.Anything, mock.Anything).
					Return([]domain.DemandAdjustment{{ID: uuid.New()}}, nil)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:     "byProduct - nil product_id",
			testType: "byProduct",
			input: ListFADsByProductInput{
				ProductID:      uuid.Nil,
				OrganizationID: uuid.New(),
			},
			setupMock:   func(m *MockDemandAdjustmentRepository) {},
			wantErr:     true,
			errContains: "product_id is required",
		},
		{
			name:     "byOrganization - success",
			testType: "byOrganization",
			input: ListFADsByOrganizationInput{
				OrganizationID: uuid.New(),
				Limit:          10,
				Offset:         0,
			},
			setupMock: func(m *MockDemandAdjustmentRepository) {
				m.On("ListByOrganization", mock.Anything, mock.Anything, 10, 0).
					Return([]domain.DemandAdjustment{{ID: uuid.New()}, {ID: uuid.New()}}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "byOrganization - nil organization_id",
			testType: "byOrganization",
			input: ListFADsByOrganizationInput{
				OrganizationID: uuid.Nil,
				Limit:          10,
				Offset:         0,
			},
			setupMock:   func(m *MockDemandAdjustmentRepository) {},
			wantErr:     true,
			errContains: "organization_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDemandAdjustmentRepository)
			tt.setupMock(mockRepo)

			useCase := NewListFADsUseCase(mockRepo)
			var result []domain.DemandAdjustment
			var err error

			if tt.testType == "byProduct" {
				result, err = useCase.ExecuteByProduct(context.Background(), tt.input.(ListFADsByProductInput))
			} else {
				result, err = useCase.ExecuteByOrganization(context.Background(), tt.input.(ListFADsByOrganizationInput))
			}

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
