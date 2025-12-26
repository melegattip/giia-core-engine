// Package events provides shared event types and utilities for NATS messaging.
package events

// Event subject constants for NATS JetStream.
// Format: {service}.{resource}.{action}
const (
	// Stream names
	StreamExecution = "EXECUTION"
	StreamDDMRP     = "DDMRP"
	StreamCatalog   = "CATALOG"
	StreamAuth      = "AUTH"
	StreamAnalytics = "ANALYTICS"
	StreamAIHub     = "AIHUB"

	// ===========================================
	// Execution Service Subjects
	// ===========================================

	// Purchase Order events
	SubjectPOCreated   = "execution.purchase_order.created"
	SubjectPOUpdated   = "execution.purchase_order.updated"
	SubjectPOReceived  = "execution.purchase_order.received"
	SubjectPOCancelled = "execution.purchase_order.cancelled"
	SubjectPOApproved  = "execution.purchase_order.approved"

	// Sales Order events
	SubjectSOCreated            = "execution.sales_order.created"
	SubjectSOUpdated            = "execution.sales_order.updated"
	SubjectSOShipped            = "execution.sales_order.shipped"
	SubjectSOCancelled          = "execution.sales_order.cancelled"
	SubjectSODeliveryNoteIssued = "execution.sales_order.delivery_note_issued"

	// Inventory events
	SubjectInventoryUpdated      = "execution.inventory.updated"
	SubjectInventoryAdjusted     = "execution.inventory.adjusted"
	SubjectInventoryTransferred  = "execution.inventory.transferred"
	SubjectInventoryBalanceAlert = "execution.inventory.balance_alert"

	// Execution alerts
	SubjectExecutionAlertCreated  = "execution.alert.created"
	SubjectExecutionAlertResolved = "execution.alert.resolved"

	// ===========================================
	// DDMRP Service Subjects
	// ===========================================

	// Buffer events
	SubjectBufferCreated        = "ddmrp.buffer.created"
	SubjectBufferUpdated        = "ddmrp.buffer.updated"
	SubjectBufferCalculated     = "ddmrp.buffer.calculated"
	SubjectBufferStatusChanged  = "ddmrp.buffer.status_changed"
	SubjectBufferAlertTriggered = "ddmrp.buffer.alert_triggered"
	SubjectBufferZoneChanged    = "ddmrp.buffer.zone_changed"

	// Demand Adjustment (FAD) events
	SubjectFADCreated = "ddmrp.fad.created"
	SubjectFADUpdated = "ddmrp.fad.updated"
	SubjectFADDeleted = "ddmrp.fad.deleted"
	SubjectFADApplied = "ddmrp.fad.applied"

	// ADU events
	SubjectADUCalculated = "ddmrp.adu.calculated"

	// ===========================================
	// Catalog Service Subjects
	// ===========================================

	SubjectProductCreated  = "catalog.product.created"
	SubjectProductUpdated  = "catalog.product.updated"
	SubjectProductDeleted  = "catalog.product.deleted"
	SubjectLocationCreated = "catalog.location.created"
	SubjectLocationUpdated = "catalog.location.updated"
	SubjectSupplierCreated = "catalog.supplier.created"
	SubjectSupplierUpdated = "catalog.supplier.updated"

	// ===========================================
	// Auth Service Subjects
	// ===========================================

	SubjectUserCreated         = "auth.user.created"
	SubjectUserUpdated         = "auth.user.updated"
	SubjectUserDeleted         = "auth.user.deleted"
	SubjectUserLoggedIn        = "auth.user.logged_in"
	SubjectUserLoggedOut       = "auth.user.logged_out"
	SubjectOrganizationCreated = "auth.organization.created"
	SubjectOrganizationUpdated = "auth.organization.updated"

	// ===========================================
	// Analytics Service Subjects
	// ===========================================

	SubjectReportGenerated   = "analytics.report.generated"
	SubjectDashboardUpdated  = "analytics.dashboard.updated"
	SubjectMetricRecorded    = "analytics.metric.recorded"
	SubjectAlertThresholdMet = "analytics.alert.threshold_met"

	// ===========================================
	// AI Intelligence Hub Subjects
	// ===========================================

	SubjectInsightGenerated      = "aihub.insight.generated"
	SubjectRecommendationCreated = "aihub.recommendation.created"
	SubjectPatternDetected       = "aihub.pattern.detected"
	SubjectAnomalyDetected       = "aihub.anomaly.detected"
)

// Consumer group names for durable subscriptions.
const (
	ConsumerExecutionService = "execution-service"
	ConsumerDDMRPService     = "ddmrp-engine-service"
	ConsumerCatalogService   = "catalog-service"
	ConsumerAnalyticsService = "analytics-service"
	ConsumerAIHubService     = "ai-intelligence-hub"
	ConsumerAuthService      = "auth-service"
)

// Source service identifiers.
const (
	SourceExecution = "execution-service"
	SourceDDMRP     = "ddmrp-engine-service"
	SourceCatalog   = "catalog-service"
	SourceAuth      = "auth-service"
	SourceAnalytics = "analytics-service"
	SourceAIHub     = "ai-intelligence-hub"
)

// Wildcard subjects for subscribing to all events from a service.
const (
	SubjectExecutionAll = "execution.>"
	SubjectDDMRPAll     = "ddmrp.>"
	SubjectCatalogAll   = "catalog.>"
	SubjectAuthAll      = "auth.>"
	SubjectAnalyticsAll = "analytics.>"
	SubjectAIHubAll     = "aihub.>"
)

// Subject patterns for specific resource types.
const (
	PatternPOAll        = "execution.purchase_order.>"
	PatternSOAll        = "execution.sales_order.>"
	PatternInventoryAll = "execution.inventory.>"
	PatternBufferAll    = "ddmrp.buffer.>"
	PatternFADAll       = "ddmrp.fad.>"
)
