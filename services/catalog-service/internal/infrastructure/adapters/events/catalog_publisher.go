package events

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/pkg/events"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
)

const (
	streamName                   = "CATALOG_EVENTS"
	productCreatedSubject        = "catalog.product.created"
	productUpdatedSubject        = "catalog.product.updated"
	productDeletedSubject        = "catalog.product.deleted"
	supplierCreatedSubject       = "catalog.supplier.created"
	supplierUpdatedSubject       = "catalog.supplier.updated"
	supplierDeletedSubject       = "catalog.supplier.deleted"
	bufferProfileAssignedSubject = "catalog.buffer_profile.assigned"
)

type catalogEventPublisher struct {
	publisher events.Publisher
	logger    logger.Logger
}

func NewCatalogEventPublisher(publisher events.Publisher, logger logger.Logger) providers.EventPublisher {
	return &catalogEventPublisher{
		publisher: publisher,
		logger:    logger,
	}
}

func (p *catalogEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	event := events.NewEvent(
		"product.created",
		"catalog-service",
		product.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"product_id":      product.ID.String(),
			"sku":             product.SKU,
			"name":            product.Name,
			"category":        product.Category,
			"unit_of_measure": product.UnitOfMeasure,
			"status":          string(product.Status),
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish product created event", logger.Tags{
			"product_id": product.ID.String(),
		})
		return err
	}

	return nil
}

func (p *catalogEventPublisher) PublishProductUpdated(ctx context.Context, product *domain.Product) error {
	event := events.NewEvent(
		"product.updated",
		"catalog-service",
		product.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"product_id": product.ID.String(),
			"sku":        product.SKU,
			"name":       product.Name,
			"category":   product.Category,
			"status":     string(product.Status),
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish product updated event", logger.Tags{
			"product_id": product.ID.String(),
		})
		return err
	}

	return nil
}

func (p *catalogEventPublisher) PublishProductDeleted(ctx context.Context, product *domain.Product) error {
	event := events.NewEvent(
		"product.deleted",
		"catalog-service",
		product.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"product_id": product.ID.String(),
			"sku":        product.SKU,
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish product deleted event", logger.Tags{
			"product_id": product.ID.String(),
		})
		return err
	}

	return nil
}

func (p *catalogEventPublisher) PublishSupplierCreated(ctx context.Context, supplier *domain.Supplier) error {
	event := events.NewEvent(
		"supplier.created",
		"catalog-service",
		supplier.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"supplier_id": supplier.ID.String(),
			"code":        supplier.Code,
			"name":        supplier.Name,
			"status":      string(supplier.Status),
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish supplier created event", logger.Tags{
			"supplier_id": supplier.ID.String(),
		})
		return err
	}

	return nil
}

func (p *catalogEventPublisher) PublishSupplierUpdated(ctx context.Context, supplier *domain.Supplier) error {
	event := events.NewEvent(
		"supplier.updated",
		"catalog-service",
		supplier.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"supplier_id": supplier.ID.String(),
			"code":        supplier.Code,
			"name":        supplier.Name,
			"status":      string(supplier.Status),
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish supplier updated event", logger.Tags{
			"supplier_id": supplier.ID.String(),
		})
		return err
	}

	return nil
}

func (p *catalogEventPublisher) PublishSupplierDeleted(ctx context.Context, supplier *domain.Supplier) error {
	event := events.NewEvent(
		"supplier.deleted",
		"catalog-service",
		supplier.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"supplier_id": supplier.ID.String(),
			"code":        supplier.Code,
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish supplier deleted event", logger.Tags{
			"supplier_id": supplier.ID.String(),
		})
		return err
	}

	return nil
}

func (p *catalogEventPublisher) PublishBufferProfileAssigned(ctx context.Context, product *domain.Product, profile *domain.BufferProfile) error {
	event := events.NewEvent(
		"buffer_profile.assigned",
		"catalog-service",
		product.OrganizationID.String(),
		time.Now(),
		map[string]interface{}{
			"product_id":   product.ID.String(),
			"sku":          product.SKU,
			"profile_id":   profile.ID.String(),
			"profile_name": profile.Name,
		},
	)

	if err := p.publisher.Publish(ctx, streamName, event); err != nil {
		p.logger.Error(ctx, err, "Failed to publish buffer profile assigned event", logger.Tags{
			"product_id": product.ID.String(),
			"profile_id": profile.ID.String(),
		})
		return err
	}

	return nil
}
