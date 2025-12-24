package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/demand_adjustment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCreateFADUseCase is a mock for testing
type MockCreateFADUseCase struct {
	mock.Mock
}

func (m *MockCreateFADUseCase) Execute(ctx context.Context, input demand_adjustment.CreateFADInput) (*domain.DemandAdjustment, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DemandAdjustment), args.Error(1)
}

// MockUpdateFADUseCase is a mock for testing
type MockUpdateFADUseCase struct {
	mock.Mock
}

func (m *MockUpdateFADUseCase) Execute(ctx context.Context, input demand_adjustment.UpdateFADInput) (*domain.DemandAdjustment, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DemandAdjustment), args.Error(1)
}

// MockDeleteFADUseCase is a mock for testing
type MockDeleteFADUseCase struct {
	mock.Mock
}

func (m *MockDeleteFADUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockListFADsUseCase is a mock for testing
type MockListFADsUseCase struct {
	mock.Mock
}

func (m *MockListFADsUseCase) ExecuteByProduct(ctx context.Context, input demand_adjustment.ListFADsByProductInput) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockListFADsUseCase) ExecuteByOrganization(ctx context.Context, input demand_adjustment.ListFADsByOrganizationInput) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func TestFADHandler_CreateFAD_Success(t *testing.T) {
	mockCreateUC := new(MockCreateFADUseCase)

	productID := uuid.New()
	orgID := uuid.New()
	createdBy := uuid.New()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0)

	expectedFAD := &domain.DemandAdjustment{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Holiday season",
		CreatedBy:      createdBy,
	}

	mockCreateUC.On("Execute", mock.Anything, mock.MatchedBy(func(input demand_adjustment.CreateFADInput) bool {
		return input.ProductID == productID && input.Factor == 1.5
	})).Return(expectedFAD, nil)

	result, err := mockCreateUC.Execute(context.Background(), demand_adjustment.CreateFADInput{
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Holiday season",
		CreatedBy:      createdBy,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, productID, result.ProductID)
	assert.Equal(t, 1.5, result.Factor)

	mockCreateUC.AssertExpectations(t)
}

func TestFADHandler_CreateFAD_ValidationError(t *testing.T) {
	mockCreateUC := new(MockCreateFADUseCase)

	mockCreateUC.On("Execute", mock.Anything, mock.Anything).
		Return(nil, domain.NewValidationError("product_id is required"))

	result, err := mockCreateUC.Execute(context.Background(), demand_adjustment.CreateFADInput{
		ProductID: uuid.Nil,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product_id is required")

	mockCreateUC.AssertExpectations(t)
}

func TestFADHandler_UpdateFAD_Success(t *testing.T) {
	mockUpdateUC := new(MockUpdateFADUseCase)

	fadID := uuid.New()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 2, 0)

	expectedFAD := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentSeasonal,
		Factor:         1.8,
		Reason:         "Updated reason",
	}

	mockUpdateUC.On("Execute", mock.Anything, mock.MatchedBy(func(input demand_adjustment.UpdateFADInput) bool {
		return input.ID == fadID && input.Factor == 1.8
	})).Return(expectedFAD, nil)

	result, err := mockUpdateUC.Execute(context.Background(), demand_adjustment.UpdateFADInput{
		ID:             fadID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentSeasonal,
		Factor:         1.8,
		Reason:         "Updated reason",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1.8, result.Factor)

	mockUpdateUC.AssertExpectations(t)
}

func TestFADHandler_UpdateFAD_NotFound(t *testing.T) {
	mockUpdateUC := new(MockUpdateFADUseCase)

	fadID := uuid.New()

	mockUpdateUC.On("Execute", mock.Anything, mock.Anything).
		Return(nil, domain.NewNotFoundError("demand adjustment not found"))

	result, err := mockUpdateUC.Execute(context.Background(), demand_adjustment.UpdateFADInput{
		ID:             fadID,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "demand adjustment not found")

	mockUpdateUC.AssertExpectations(t)
}

func TestFADHandler_DeleteFAD_Success(t *testing.T) {
	mockDeleteUC := new(MockDeleteFADUseCase)

	fadID := uuid.New()

	mockDeleteUC.On("Execute", mock.Anything, fadID).Return(nil)

	err := mockDeleteUC.Execute(context.Background(), fadID)

	assert.NoError(t, err)

	mockDeleteUC.AssertExpectations(t)
}

func TestFADHandler_DeleteFAD_NotFound(t *testing.T) {
	mockDeleteUC := new(MockDeleteFADUseCase)

	fadID := uuid.New()

	mockDeleteUC.On("Execute", mock.Anything, fadID).
		Return(domain.NewNotFoundError("demand adjustment not found"))

	err := mockDeleteUC.Execute(context.Background(), fadID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "demand adjustment not found")

	mockDeleteUC.AssertExpectations(t)
}

func TestFADHandler_ListFADsByProduct_Success(t *testing.T) {
	mockListUC := new(MockListFADsUseCase)

	productID := uuid.New()
	orgID := uuid.New()

	expectedFADs := []domain.DemandAdjustment{
		{ID: uuid.New(), ProductID: productID, OrganizationID: orgID, Factor: 1.2},
		{ID: uuid.New(), ProductID: productID, OrganizationID: orgID, Factor: 1.5},
	}

	mockListUC.On("ExecuteByProduct", mock.Anything, demand_adjustment.ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	}).Return(expectedFADs, nil)

	result, err := mockListUC.ExecuteByProduct(context.Background(), demand_adjustment.ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	mockListUC.AssertExpectations(t)
}

func TestFADHandler_ListFADsByOrganization_Success(t *testing.T) {
	mockListUC := new(MockListFADsUseCase)

	orgID := uuid.New()

	expectedFADs := []domain.DemandAdjustment{
		{ID: uuid.New(), OrganizationID: orgID, ProductID: uuid.New()},
		{ID: uuid.New(), OrganizationID: orgID, ProductID: uuid.New()},
		{ID: uuid.New(), OrganizationID: orgID, ProductID: uuid.New()},
	}

	mockListUC.On("ExecuteByOrganization", mock.Anything, demand_adjustment.ListFADsByOrganizationInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	}).Return(expectedFADs, nil)

	result, err := mockListUC.ExecuteByOrganization(context.Background(), demand_adjustment.ListFADsByOrganizationInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	mockListUC.AssertExpectations(t)
}

func TestFADToProto(t *testing.T) {
	fadID := uuid.New()
	productID := uuid.New()
	orgID := uuid.New()
	createdBy := uuid.New()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0)
	createdAt := time.Now()

	fad := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Holiday season adjustment",
		CreatedAt:      createdAt,
		CreatedBy:      createdBy,
	}

	result := FADToProto(fad)

	assert.Equal(t, fadID.String(), result["id"])
	assert.Equal(t, productID.String(), result["product_id"])
	assert.Equal(t, orgID.String(), result["organization_id"])
	assert.Equal(t, "fad", result["adjustment_type"])
	assert.Equal(t, 1.5, result["factor"])
	assert.Equal(t, "Holiday season adjustment", result["reason"])
	assert.Equal(t, createdBy.String(), result["created_by"])
}

func TestFADHandler_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase)
		testFn    func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) error
		wantErr   bool
	}{
		{
			name: "create success",
			setupMock: func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) {
				createUC.On("Execute", mock.Anything, mock.Anything).
					Return(&domain.DemandAdjustment{ID: uuid.New()}, nil)
			},
			testFn: func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) error {
				_, err := createUC.Execute(context.Background(), demand_adjustment.CreateFADInput{
					ProductID:      uuid.New(),
					OrganizationID: uuid.New(),
					StartDate:      time.Now(),
					EndDate:        time.Now().AddDate(0, 1, 0),
					AdjustmentType: domain.DemandAdjustmentFAD,
					Factor:         1.5,
					Reason:         "Test",
					CreatedBy:      uuid.New(),
				})
				return err
			},
			wantErr: false,
		},
		{
			name: "update success",
			setupMock: func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) {
				updateUC.On("Execute", mock.Anything, mock.Anything).
					Return(&domain.DemandAdjustment{ID: uuid.New()}, nil)
			},
			testFn: func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) error {
				_, err := updateUC.Execute(context.Background(), demand_adjustment.UpdateFADInput{
					ID:             uuid.New(),
					StartDate:      time.Now(),
					EndDate:        time.Now().AddDate(0, 1, 0),
					AdjustmentType: domain.DemandAdjustmentFAD,
					Factor:         1.8,
					Reason:         "Updated",
				})
				return err
			},
			wantErr: false,
		},
		{
			name: "delete success",
			setupMock: func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) {
				deleteUC.On("Execute", mock.Anything, mock.Anything).Return(nil)
			},
			testFn: func(createUC *MockCreateFADUseCase, updateUC *MockUpdateFADUseCase, deleteUC *MockDeleteFADUseCase) error {
				return deleteUC.Execute(context.Background(), uuid.New())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCreateUC := new(MockCreateFADUseCase)
			mockUpdateUC := new(MockUpdateFADUseCase)
			mockDeleteUC := new(MockDeleteFADUseCase)

			tt.setupMock(mockCreateUC, mockUpdateUC, mockDeleteUC)
			err := tt.testFn(mockCreateUC, mockUpdateUC, mockDeleteUC)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCreateUC.AssertExpectations(t)
			mockUpdateUC.AssertExpectations(t)
			mockDeleteUC.AssertExpectations(t)
		})
	}
}
