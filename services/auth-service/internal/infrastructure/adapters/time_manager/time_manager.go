package time_manager

import "time"

type TimeManager struct{}

func NewTimeManager() *TimeManager {
	return &TimeManager{}
}

func (t *TimeManager) Now() time.Time {
	return time.Now()
}
