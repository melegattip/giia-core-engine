package time

import (
	"time"
)

type TimeManager interface {
	Now() time.Time
	FormatToISO8601(date time.Time) string
	FormatToOffset(date time.Time) string
	StringToUTC(dateString string) (time.Time, error)
	StringYearMonthDayToUTC(date string) (time.Time, error)
}

type RealTimeManager struct{}

func NewRealTimeManager() *RealTimeManager {
	return &RealTimeManager{}
}

func (tm *RealTimeManager) Now() time.Time {
	return time.Now().UTC()
}

func (tm *RealTimeManager) FormatToISO8601(date time.Time) string {
	return date.UTC().Format(time.RFC3339)
}

func (tm *RealTimeManager) FormatToOffset(date time.Time) string {
	return date.Format(time.RFC3339)
}

func (tm *RealTimeManager) StringToUTC(dateString string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime.UTC(), nil
}

func (tm *RealTimeManager) StringYearMonthDayToUTC(date string) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime.UTC(), nil
}
