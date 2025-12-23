package bufferProfile

import (
	"context"
	"testing"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBufferProfileRepository struct {
	mock.Mock
}

func (m *MockBufferProfileRepository) Create(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockBufferProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BufferProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BufferProfile), args.Error(1)
}

func (m *MockBufferProfileRepository) GetByName(ctx context.Context, name string) (*domain.BufferProfile, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BufferProfile), args.Error(1)
}

func (m *MockBufferProfileRepository) Update(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockBufferProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBufferProfileRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.BufferProfile, int64, error) {
	args := m.Called(ctx, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.BufferProfile), args.Get(1).(int64), args.Error(2)
}

type MockBufferEventPublisher struct {
	mock.Mock
}

func (m *MockBufferEventPublisher) PublishBufferProfileCreated(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishBufferProfileUpdated(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishBufferProfileDeleted(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishBufferProfileAssigned(ctx context.Context, product *domain.Product, profile *domain.BufferProfile) error {
	args := m.Called(ctx, product, profile)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishProductUpdated(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishProductDeleted(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishSupplierCreated(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishSupplierUpdated(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockBufferEventPublisher) PublishSupplierDeleted(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

var _ providers.BufferProfileRepository = (*MockBufferProfileRepository)(nil)
var _ providers.EventPublisher = (*MockBufferEventPublisher)(nil)

// CREATE TESTS
func TestCreateBufferProfileUseCase_Execute_WithValidData_ReturnsProfile(t *testing.T) {
	givenOrgID := uuid.New()
	givenRequest := &CreateBufferProfileRequest{
		Name:               "Standard Buffer",
		Description:        "Standard buffer profile",
		LeadTimeFactor:     1.5,
		VariabilityFactor:  1.2,
		TargetServiceLevel: 95,
	}

	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(bp *domain.BufferProfile) bool {
		return bp.Name == "Standard Buffer" && bp.LeadTimeFactor == 1.5
	})).Return(nil)
	mockPublisher.On("PublishBufferProfileCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewCreateBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	profile, err := useCase.Execute(ctx, givenRequest)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "Standard Buffer", profile.Name)
	assert.Equal(t, 1.5, profile.LeadTimeFactor)
	assert.Equal(t, 1.2, profile.VariabilityFactor)
	assert.Equal(t, 95, profile.TargetServiceLevel)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreateBufferProfileUseCase_Execute_WithDefaultServiceLevel_Sets95(t *testing.T) {
	givenOrgID := uuid.New()
	givenRequest := &CreateBufferProfileRequest{
		Name:               "Standard Buffer",
		Description:        "Standard buffer profile",
		LeadTimeFactor:     1.5,
		VariabilityFactor:  1.2,
		TargetServiceLevel: 0,
	}

	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(bp *domain.BufferProfile) bool {
		return bp.TargetServiceLevel == 95
	})).Return(nil)
	mockPublisher.On("PublishBufferProfileCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewCreateBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	profile, err := useCase.Execute(ctx, givenRequest)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, 95, profile.TargetServiceLevel)
	mockRepo.AssertExpectations(t)
}

func TestCreateBufferProfileUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewCreateBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	profile, err := useCase.Execute(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateBufferProfileUseCase_Execute_WithMissingOrgID_ReturnsError(t *testing.T) {
	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewCreateBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	request := &CreateBufferProfileRequest{
		Name:               "Standard Buffer",
		LeadTimeFactor:     1.5,
		VariabilityFactor:  1.2,
		TargetServiceLevel: 95,
	}

	profile, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "Create")
}

// DELETE TESTS
func TestDeleteBufferProfileUseCase_Execute_WithValidID_DeletesProfile(t *testing.T) {
	givenProfileID := uuid.New()
	givenOrgID := uuid.New()
	givenProfile := &domain.BufferProfile{
		ID:                 givenProfileID,
		Name:               "Test Profile",
		LeadTimeFactor:     1.5,
		VariabilityFactor:  1.2,
		TargetServiceLevel: 95,
		OrganizationID:     givenOrgID,
	}

	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProfileID).Return(givenProfile, nil)
	mockRepo.On("Delete", mock.Anything, givenProfileID).Return(nil)
	mockPublisher.On("PublishBufferProfileDeleted", mock.Anything, mock.MatchedBy(func(bp *domain.BufferProfile) bool {
		return bp.ID == givenProfileID
	})).Return(nil)

	useCase := NewDeleteBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	err := useCase.Execute(ctx, givenProfileID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestDeleteBufferProfileUseCase_Execute_WithNilID_ReturnsError(t *testing.T) {
	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewDeleteBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	err := useCase.Execute(ctx, uuid.Nil)

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Delete")
}

// GET TESTS
func TestGetBufferProfileUseCase_Execute_WithValidID_ReturnsProfile(t *testing.T) {
	givenProfileID := uuid.New()
	givenOrgID := uuid.New()
	givenProfile := &domain.BufferProfile{
		ID:                 givenProfileID,
		Name:               "Test Profile",
		LeadTimeFactor:     1.5,
		VariabilityFactor:  1.2,
		TargetServiceLevel: 95,
		OrganizationID:     givenOrgID,
	}

	mockRepo := new(MockBufferProfileRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProfileID).Return(givenProfile, nil)

	useCase := NewGetBufferProfileUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	profile, err := useCase.Execute(ctx, givenProfileID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, givenProfileID, profile.ID)
	assert.Equal(t, "Test Profile", profile.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetBufferProfileUseCase_Execute_WithNilID_ReturnsError(t *testing.T) {
	mockRepo := new(MockBufferProfileRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewGetBufferProfileUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	profile, err := useCase.Execute(ctx, uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

// LIST TESTS
func TestListBufferProfilesUseCase_Execute_WithValidRequest_ReturnsProfiles(t *testing.T) {
	givenOrgID := uuid.New()
	givenProfiles := []*domain.BufferProfile{
		{
			ID:                 uuid.New(),
			Name:               "Profile 1",
			LeadTimeFactor:     1.5,
			VariabilityFactor:  1.2,
			TargetServiceLevel: 95,
			OrganizationID:     givenOrgID,
		},
		{
			ID:                 uuid.New(),
			Name:               "Profile 2",
			LeadTimeFactor:     2.0,
			VariabilityFactor:  1.5,
			TargetServiceLevel: 90,
			OrganizationID:     givenOrgID,
		},
	}

	mockRepo := new(MockBufferProfileRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenProfiles, int64(2), nil)

	useCase := NewListBufferProfilesUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListBufferProfilesRequest{
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response.BufferProfiles))
	assert.Equal(t, int64(2), response.Total)
	mockRepo.AssertExpectations(t)
}

func TestListBufferProfilesUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockBufferProfileRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewListBufferProfilesUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	response, err := useCase.Execute(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "List")
}

// UPDATE TESTS
func TestUpdateBufferProfileUseCase_Execute_WithValidData_ReturnsUpdatedProfile(t *testing.T) {
	givenProfileID := uuid.New()
	givenOrgID := uuid.New()
	givenExistingProfile := &domain.BufferProfile{
		ID:                 givenProfileID,
		Name:               "Old Name",
		LeadTimeFactor:     1.5,
		VariabilityFactor:  1.2,
		TargetServiceLevel: 95,
		OrganizationID:     givenOrgID,
	}

	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProfileID).Return(givenExistingProfile, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(bp *domain.BufferProfile) bool {
		return bp.ID == givenProfileID && bp.Name == "New Name"
	})).Return(nil)
	mockPublisher.On("PublishBufferProfileUpdated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewUpdateBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &UpdateBufferProfileRequest{
		ID:                 givenProfileID,
		Name:               "New Name",
		LeadTimeFactor:     2.0,
		VariabilityFactor:  1.5,
		TargetServiceLevel: 98,
	}

	profile, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "New Name", profile.Name)
	assert.Equal(t, 2.0, profile.LeadTimeFactor)
	assert.Equal(t, 1.5, profile.VariabilityFactor)
	assert.Equal(t, 98, profile.TargetServiceLevel)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestUpdateBufferProfileUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockBufferProfileRepository)
	mockPublisher := new(MockBufferEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewUpdateBufferProfileUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	profile, err := useCase.Execute(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}
