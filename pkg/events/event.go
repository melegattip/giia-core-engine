package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/errors"
)

type Event struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Source         string                 `json:"source"`
	OrganizationID string                 `json:"organization_id"`
	Timestamp      time.Time              `json:"timestamp"`
	SchemaVersion  string                 `json:"schema_version"`
	Data           map[string]interface{} `json:"data"`
}

func NewEvent(eventType, source, organizationID string, timestamp time.Time, data map[string]interface{}) *Event {
	return &Event{
		ID:             uuid.New().String(),
		Type:           eventType,
		Source:         source,
		OrganizationID: organizationID,
		Timestamp:      timestamp,
		SchemaVersion:  "1.0",
		Data:           data,
	}
}

func (e *Event) Validate() error {
	if e.ID == "" {
		return errors.NewBadRequest("event ID is required")
	}

	if e.Type == "" {
		return errors.NewBadRequest("event type is required")
	}

	if e.Source == "" {
		return errors.NewBadRequest("event source is required")
	}

	if e.OrganizationID == "" {
		return errors.NewBadRequest("organization_id is required")
	}

	if e.Timestamp.IsZero() {
		return errors.NewBadRequest("timestamp is required")
	}

	if e.SchemaVersion == "" {
		return errors.NewBadRequest("schema_version is required")
	}

	return nil
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func FromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, errors.NewBadRequest("failed to unmarshal event")
	}
	return &event, nil
}
