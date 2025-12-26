package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferCreatedEvent_Marshal(t *testing.T) {
	event := BufferCreatedEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		ProfileType:    "standard",
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded BufferCreatedEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.BufferID, decoded.BufferID)
	assert.Equal(t, event.ProductID, decoded.ProductID)
	assert.Equal(t, event.IsActive, decoded.IsActive)
}

func TestBufferCalculatedEvent_Marshal(t *testing.T) {
	event := BufferCalculatedEvent{
		BufferID:        "buf-123",
		OrganizationID:  "org-123",
		ProductID:       "prod-123",
		LocationID:      "loc-123",
		DLT:             14,
		ADU:             100.0,
		RedZone:         200.0,
		YellowZone:      600.0,
		GreenZone:       400.0,
		CPD:             0.75,
		TOG:             1200.0,
		TOR:             200.0,
		TOY:             800.0,
		OnHandQty:       500.0,
		OpenPOQty:       300.0,
		OpenSOQty:       100.0,
		NetFlowPosition: 700.0,
		CalculatedAt:    time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded BufferCalculatedEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.BufferID, decoded.BufferID)
	assert.Equal(t, event.ADU, decoded.ADU)
	assert.Equal(t, event.TOG, decoded.TOG)
	assert.Equal(t, event.NetFlowPosition, decoded.NetFlowPosition)
}

func TestBufferStatusChangedEvent_Marshal(t *testing.T) {
	event := BufferStatusChangedEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		OldZone:        "yellow",
		NewZone:        "red",
		OldNFP:         400.0,
		NewNFP:         150.0,
		AlertLevel:     "warning",
		TOG:            1200.0,
		TOY:            800.0,
		TOR:            200.0,
		ChangedAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded BufferStatusChangedEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.BufferID, decoded.BufferID)
	assert.Equal(t, event.OldZone, decoded.OldZone)
	assert.Equal(t, event.NewZone, decoded.NewZone)
	assert.Equal(t, event.AlertLevel, decoded.AlertLevel)
}

func TestBufferAlertTriggeredEvent_Marshal(t *testing.T) {
	event := BufferAlertTriggeredEvent{
		BufferID:              "buf-123",
		OrganizationID:        "org-123",
		ProductID:             "prod-123",
		ProductSKU:            "SKU-001",
		ProductName:           "Test Product",
		LocationID:            "loc-123",
		LocationName:          "Main Warehouse",
		AlertType:             "low_stock",
		AlertLevel:            "critical",
		Zone:                  "red",
		NFP:                   100.0,
		TOG:                   1200.0,
		TOY:                   800.0,
		TOR:                   200.0,
		ReplenishmentQty:      1100.0,
		SuggestedOrderQty:     1100.0,
		RecommendedSupplierID: "sup-123",
		ExpectedLeadTimeDays:  7,
		Message:               "Buffer is in critical red zone",
		Metadata:              map[string]string{"priority": "high"},
		TriggeredAt:           time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded BufferAlertTriggeredEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.BufferID, decoded.BufferID)
	assert.Equal(t, event.AlertType, decoded.AlertType)
	assert.Equal(t, event.ReplenishmentQty, decoded.ReplenishmentQty)
	assert.Equal(t, event.Metadata["priority"], decoded.Metadata["priority"])
}

func TestFADCreatedEvent_Marshal(t *testing.T) {
	event := FADCreatedEvent{
		FADID:          "fad-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		AdjustmentType: "seasonal",
		Factor:         1.5,
		StartDate:      time.Now().UTC(),
		EndDate:        time.Now().UTC().Add(30 * 24 * time.Hour),
		Reason:         "Holiday season",
		CreatedBy:      "user-123",
		CreatedAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded FADCreatedEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.FADID, decoded.FADID)
	assert.Equal(t, event.Factor, decoded.Factor)
	assert.Equal(t, event.AdjustmentType, decoded.AdjustmentType)
}

func TestADUCalculatedEvent_Marshal(t *testing.T) {
	event := ADUCalculatedEvent{
		BufferID:          "buf-123",
		OrganizationID:    "org-123",
		ProductID:         "prod-123",
		LocationID:        "loc-123",
		ADU:               150.0,
		PreviousADU:       120.0,
		CalculationMethod: "weighted_average",
		PeriodDays:        90,
		DataPointCount:    90,
		CalculatedAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded ADUCalculatedEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.BufferID, decoded.BufferID)
	assert.Equal(t, event.ADU, decoded.ADU)
	assert.Equal(t, event.PreviousADU, decoded.PreviousADU)
	assert.Equal(t, event.PeriodDays, decoded.PeriodDays)
}

func TestSubjectConstants(t *testing.T) {
	// Verify subject naming convention
	assert.Equal(t, "ddmrp.buffer.created", SubjectBufferCreated)
	assert.Equal(t, "ddmrp.buffer.calculated", SubjectBufferCalculated)
	assert.Equal(t, "ddmrp.buffer.status_changed", SubjectBufferStatusChanged)
	assert.Equal(t, "ddmrp.buffer.alert_triggered", SubjectBufferAlertTriggered)
	assert.Equal(t, "ddmrp.fad.created", SubjectFADCreated)
	assert.Equal(t, "ddmrp.fad.updated", SubjectFADUpdated)
	assert.Equal(t, "ddmrp.fad.deleted", SubjectFADDeleted)
	assert.Equal(t, "ddmrp.adu.calculated", SubjectADUCalculated)
}

func TestTypeConstants(t *testing.T) {
	assert.Equal(t, "buffer.created", TypeBufferCreated)
	assert.Equal(t, "buffer.calculated", TypeBufferCalculated)
	assert.Equal(t, "buffer.status_changed", TypeBufferStatusChanged)
	assert.Equal(t, "buffer.alert_triggered", TypeBufferAlertTriggered)
	assert.Equal(t, "fad.created", TypeFADCreated)
	assert.Equal(t, "fad.updated", TypeFADUpdated)
	assert.Equal(t, "fad.deleted", TypeFADDeleted)
	assert.Equal(t, "adu.calculated", TypeADUCalculated)
}

func TestZoneConstants(t *testing.T) {
	assert.Equal(t, "red", ZoneRed)
	assert.Equal(t, "yellow", ZoneYellow)
	assert.Equal(t, "green", ZoneGreen)
}

func TestAlertLevelConstants(t *testing.T) {
	assert.Equal(t, "critical", AlertLevelCritical)
	assert.Equal(t, "warning", AlertLevelWarning)
	assert.Equal(t, "info", AlertLevelInfo)
}

func TestBufferZoneChangedEvent_Marshal(t *testing.T) {
	event := BufferZoneChangedEvent{
		BufferStatusChangedEvent: BufferStatusChangedEvent{
			BufferID:       "buf-123",
			OrganizationID: "org-123",
			ProductID:      "prod-123",
			LocationID:     "loc-123",
			OldZone:        "green",
			NewZone:        "yellow",
			NewNFP:         650.0,
			AlertLevel:     "info",
			TOG:            1200.0,
			TOY:            800.0,
			TOR:            200.0,
			ChangedAt:      time.Now().UTC(),
		},
		TransitionReason: "Increased demand consumption",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded BufferZoneChangedEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.BufferID, decoded.BufferID)
	assert.Equal(t, event.OldZone, decoded.OldZone)
	assert.Equal(t, event.NewZone, decoded.NewZone)
	assert.Equal(t, event.TransitionReason, decoded.TransitionReason)
}
