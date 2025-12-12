package time

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type MockTimeManager struct {
	mock.Mock
}

func NewMockTimeManager() *MockTimeManager {
	return &MockTimeManager{}
}

func (m *MockTimeManager) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockTimeManager) FormatToISO8601(date time.Time) string {
	args := m.Called(date)
	return args.String(0)
}

func (m *MockTimeManager) FormatToOffset(date time.Time) string {
	args := m.Called(date)
	return args.String(0)
}

func (m *MockTimeManager) StringToUTC(dateString string) (time.Time, error) {
	args := m.Called(dateString)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockTimeManager) StringYearMonthDayToUTC(date string) (time.Time, error) {
	args := m.Called(date)
	return args.Get(0).(time.Time), args.Error(1)
}
