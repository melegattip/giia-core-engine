package providers

import "time"

type TimeManager interface {
	Now() time.Time
}
