package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Source         string                 `json:"source"`
	OrganizationID string                 `json:"organization_id"`
	Timestamp      time.Time              `json:"timestamp"`
	Data           map[string]interface{} `json:"data"`
}

func NewEvent(eventType, source, organizationID string, data map[string]interface{}) *Event {
	return &Event{
		ID:             uuid.New().String(),
		Type:           eventType,
		Source:         source,
		OrganizationID: organizationID,
		Timestamp:      time.Now().UTC(),
		Data:           data,
	}
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func FromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
