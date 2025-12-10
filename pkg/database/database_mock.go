package database

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) Connect(ctx context.Context, config *Config) (*gorm.DB, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *DatabaseMock) HealthCheck(ctx context.Context, db *gorm.DB) error {
	args := m.Called(ctx, db)
	return args.Error(0)
}

func (m *DatabaseMock) Close(db *gorm.DB) error {
	args := m.Called(db)
	return args.Error(0)
}
