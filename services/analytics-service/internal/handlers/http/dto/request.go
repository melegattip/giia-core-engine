// Package dto provides Data Transfer Objects for HTTP handlers.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// KPIRequest represents the base request for KPI endpoints.
type KPIRequest struct {
	OrganizationID uuid.UUID  `json:"organization_id"`
	SnapshotDate   *time.Time `json:"snapshot_date,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// ImmobilizedKPIRequest extends KPIRequest with threshold.
type ImmobilizedKPIRequest struct {
	KPIRequest
	ThresholdYears int `json:"threshold_years"`
}

// BufferAnalyticsRequest represents the request for buffer analytics.
type BufferAnalyticsRequest struct {
	OrganizationID uuid.UUID  `json:"organization_id"`
	ProductID      *uuid.UUID `json:"product_id,omitempty"`
	Date           *time.Time `json:"date,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// SyncBufferRequest represents the request for syncing buffer data.
type SyncBufferRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Date           time.Time `json:"date"`
}
