package events

import (
	"context"
	"encoding/json"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
)

type NATSPublisher struct {
	enabled bool
}

func NewNATSPublisher() *NATSPublisher {
	return &NATSPublisher{
		enabled: false,
	}
}

func (p *NATSPublisher) PublishBufferCalculated(ctx context.Context, buffer *domain.Buffer) error {
	if !p.enabled {
		return nil
	}

	event := map[string]interface{}{
		"event_type":   "buffer.calculated",
		"buffer_id":    buffer.ID.String(),
		"product_id":   buffer.ProductID.String(),
		"organization": buffer.OrganizationID.String(),
		"cpd":          buffer.CPD,
		"zones": map[string]float64{
			"red":    buffer.RedZone,
			"yellow": buffer.YellowZone,
			"green":  buffer.GreenZone,
		},
	}

	_, err := json.Marshal(event)
	return err
}

func (p *NATSPublisher) PublishBufferStatusChanged(ctx context.Context, buffer *domain.Buffer, oldZone domain.ZoneType) error {
	if !p.enabled {
		return nil
	}

	event := map[string]interface{}{
		"event_type":   "buffer.status_changed",
		"buffer_id":    buffer.ID.String(),
		"product_id":   buffer.ProductID.String(),
		"organization": buffer.OrganizationID.String(),
		"old_zone":     oldZone,
		"new_zone":     buffer.Zone,
		"alert_level":  buffer.AlertLevel,
	}

	_, err := json.Marshal(event)
	return err
}

func (p *NATSPublisher) PublishBufferAlertTriggered(ctx context.Context, buffer *domain.Buffer) error {
	if !p.enabled {
		return nil
	}

	event := map[string]interface{}{
		"event_type":   "buffer.alert_triggered",
		"buffer_id":    buffer.ID.String(),
		"product_id":   buffer.ProductID.String(),
		"organization": buffer.OrganizationID.String(),
		"alert_level":  buffer.AlertLevel,
		"zone":         buffer.Zone,
		"nfp":          buffer.NetFlowPosition,
	}

	_, err := json.Marshal(event)
	return err
}

func (p *NATSPublisher) PublishFADCreated(ctx context.Context, fad *domain.DemandAdjustment) error {
	if !p.enabled {
		return nil
	}

	event := map[string]interface{}{
		"event_type":   "fad.created",
		"fad_id":       fad.ID.String(),
		"product_id":   fad.ProductID.String(),
		"organization": fad.OrganizationID.String(),
		"factor":       fad.Factor,
		"type":         fad.AdjustmentType,
	}

	_, err := json.Marshal(event)
	return err
}

func (p *NATSPublisher) PublishFADUpdated(ctx context.Context, fad *domain.DemandAdjustment) error {
	if !p.enabled {
		return nil
	}

	event := map[string]interface{}{
		"event_type":   "fad.updated",
		"fad_id":       fad.ID.String(),
		"product_id":   fad.ProductID.String(),
		"organization": fad.OrganizationID.String(),
		"factor":       fad.Factor,
	}

	_, err := json.Marshal(event)
	return err
}

func (p *NATSPublisher) PublishFADDeleted(ctx context.Context, fadID string) error {
	if !p.enabled {
		return nil
	}

	event := map[string]interface{}{
		"event_type": "fad.deleted",
		"fad_id":     fadID,
	}

	_, err := json.Marshal(event)
	return err
}
