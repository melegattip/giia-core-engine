package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamNames(t *testing.T) {
	assert.Equal(t, "EXECUTION", StreamExecution)
	assert.Equal(t, "DDMRP", StreamDDMRP)
	assert.Equal(t, "CATALOG", StreamCatalog)
	assert.Equal(t, "AUTH", StreamAuth)
	assert.Equal(t, "ANALYTICS", StreamAnalytics)
	assert.Equal(t, "AIHUB", StreamAIHub)
}

func TestExecutionSubjects(t *testing.T) {
	// Purchase Order subjects
	assert.Equal(t, "execution.purchase_order.created", SubjectPOCreated)
	assert.Equal(t, "execution.purchase_order.updated", SubjectPOUpdated)
	assert.Equal(t, "execution.purchase_order.received", SubjectPOReceived)
	assert.Equal(t, "execution.purchase_order.cancelled", SubjectPOCancelled)
	assert.Equal(t, "execution.purchase_order.approved", SubjectPOApproved)

	// Sales Order subjects
	assert.Equal(t, "execution.sales_order.created", SubjectSOCreated)
	assert.Equal(t, "execution.sales_order.updated", SubjectSOUpdated)
	assert.Equal(t, "execution.sales_order.shipped", SubjectSOShipped)
	assert.Equal(t, "execution.sales_order.cancelled", SubjectSOCancelled)
	assert.Equal(t, "execution.sales_order.delivery_note_issued", SubjectSODeliveryNoteIssued)

	// Inventory subjects
	assert.Equal(t, "execution.inventory.updated", SubjectInventoryUpdated)
	assert.Equal(t, "execution.inventory.adjusted", SubjectInventoryAdjusted)
	assert.Equal(t, "execution.inventory.transferred", SubjectInventoryTransferred)
	assert.Equal(t, "execution.inventory.balance_alert", SubjectInventoryBalanceAlert)

	// Alert subjects
	assert.Equal(t, "execution.alert.created", SubjectExecutionAlertCreated)
	assert.Equal(t, "execution.alert.resolved", SubjectExecutionAlertResolved)
}

func TestDDMRPSubjects(t *testing.T) {
	// Buffer subjects
	assert.Equal(t, "ddmrp.buffer.created", SubjectBufferCreated)
	assert.Equal(t, "ddmrp.buffer.updated", SubjectBufferUpdated)
	assert.Equal(t, "ddmrp.buffer.calculated", SubjectBufferCalculated)
	assert.Equal(t, "ddmrp.buffer.status_changed", SubjectBufferStatusChanged)
	assert.Equal(t, "ddmrp.buffer.alert_triggered", SubjectBufferAlertTriggered)
	assert.Equal(t, "ddmrp.buffer.zone_changed", SubjectBufferZoneChanged)

	// FAD subjects
	assert.Equal(t, "ddmrp.fad.created", SubjectFADCreated)
	assert.Equal(t, "ddmrp.fad.updated", SubjectFADUpdated)
	assert.Equal(t, "ddmrp.fad.deleted", SubjectFADDeleted)
	assert.Equal(t, "ddmrp.fad.applied", SubjectFADApplied)

	// ADU subject
	assert.Equal(t, "ddmrp.adu.calculated", SubjectADUCalculated)
}

func TestCatalogSubjects(t *testing.T) {
	assert.Equal(t, "catalog.product.created", SubjectProductCreated)
	assert.Equal(t, "catalog.product.updated", SubjectProductUpdated)
	assert.Equal(t, "catalog.product.deleted", SubjectProductDeleted)
	assert.Equal(t, "catalog.location.created", SubjectLocationCreated)
	assert.Equal(t, "catalog.location.updated", SubjectLocationUpdated)
	assert.Equal(t, "catalog.supplier.created", SubjectSupplierCreated)
	assert.Equal(t, "catalog.supplier.updated", SubjectSupplierUpdated)
}

func TestAuthSubjects(t *testing.T) {
	assert.Equal(t, "auth.user.created", SubjectUserCreated)
	assert.Equal(t, "auth.user.updated", SubjectUserUpdated)
	assert.Equal(t, "auth.user.deleted", SubjectUserDeleted)
	assert.Equal(t, "auth.user.logged_in", SubjectUserLoggedIn)
	assert.Equal(t, "auth.user.logged_out", SubjectUserLoggedOut)
	assert.Equal(t, "auth.organization.created", SubjectOrganizationCreated)
	assert.Equal(t, "auth.organization.updated", SubjectOrganizationUpdated)
}

func TestAnalyticsSubjects(t *testing.T) {
	assert.Equal(t, "analytics.report.generated", SubjectReportGenerated)
	assert.Equal(t, "analytics.dashboard.updated", SubjectDashboardUpdated)
	assert.Equal(t, "analytics.metric.recorded", SubjectMetricRecorded)
	assert.Equal(t, "analytics.alert.threshold_met", SubjectAlertThresholdMet)
}

func TestAIHubSubjects(t *testing.T) {
	assert.Equal(t, "aihub.insight.generated", SubjectInsightGenerated)
	assert.Equal(t, "aihub.recommendation.created", SubjectRecommendationCreated)
	assert.Equal(t, "aihub.pattern.detected", SubjectPatternDetected)
	assert.Equal(t, "aihub.anomaly.detected", SubjectAnomalyDetected)
}

func TestConsumerNames(t *testing.T) {
	assert.Equal(t, "execution-service", ConsumerExecutionService)
	assert.Equal(t, "ddmrp-engine-service", ConsumerDDMRPService)
	assert.Equal(t, "catalog-service", ConsumerCatalogService)
	assert.Equal(t, "analytics-service", ConsumerAnalyticsService)
	assert.Equal(t, "ai-intelligence-hub", ConsumerAIHubService)
	assert.Equal(t, "auth-service", ConsumerAuthService)
}

func TestSourceIdentifiers(t *testing.T) {
	assert.Equal(t, "execution-service", SourceExecution)
	assert.Equal(t, "ddmrp-engine-service", SourceDDMRP)
	assert.Equal(t, "catalog-service", SourceCatalog)
	assert.Equal(t, "auth-service", SourceAuth)
	assert.Equal(t, "analytics-service", SourceAnalytics)
	assert.Equal(t, "ai-intelligence-hub", SourceAIHub)
}

func TestWildcardSubjects(t *testing.T) {
	assert.Equal(t, "execution.>", SubjectExecutionAll)
	assert.Equal(t, "ddmrp.>", SubjectDDMRPAll)
	assert.Equal(t, "catalog.>", SubjectCatalogAll)
	assert.Equal(t, "auth.>", SubjectAuthAll)
	assert.Equal(t, "analytics.>", SubjectAnalyticsAll)
	assert.Equal(t, "aihub.>", SubjectAIHubAll)
}

func TestSubjectPatterns(t *testing.T) {
	assert.Equal(t, "execution.purchase_order.>", PatternPOAll)
	assert.Equal(t, "execution.sales_order.>", PatternSOAll)
	assert.Equal(t, "execution.inventory.>", PatternInventoryAll)
	assert.Equal(t, "ddmrp.buffer.>", PatternBufferAll)
	assert.Equal(t, "ddmrp.fad.>", PatternFADAll)
}
