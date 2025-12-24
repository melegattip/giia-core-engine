// Package handlers provides HTTP handlers for the Analytics Service.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http/cache"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http/dto"
)

// KPIHandler handles HTTP requests for KPI operations.
type KPIHandler struct {
	kpiRepo            providers.KPIRepository
	diiUseCase         *kpi.CalculateDaysInInventoryUseCase
	immobilizedUseCase *kpi.CalculateImmobilizedInventoryUseCase
	rotationUseCase    *kpi.CalculateInventoryRotationUseCase
	syncBufferUseCase  *kpi.SyncBufferAnalyticsUseCase
	cache              *cache.Cache
}

// NewKPIHandler creates a new KPI handler.
func NewKPIHandler(
	kpiRepo providers.KPIRepository,
	diiUseCase *kpi.CalculateDaysInInventoryUseCase,
	immobilizedUseCase *kpi.CalculateImmobilizedInventoryUseCase,
	rotationUseCase *kpi.CalculateInventoryRotationUseCase,
	syncBufferUseCase *kpi.SyncBufferAnalyticsUseCase,
) *KPIHandler {
	return &KPIHandler{
		kpiRepo:            kpiRepo,
		diiUseCase:         diiUseCase,
		immobilizedUseCase: immobilizedUseCase,
		rotationUseCase:    rotationUseCase,
		syncBufferUseCase:  syncBufferUseCase,
		cache:              cache.New(5 * time.Minute),
	}
}

// GetDaysInInventory handles GET /api/v1/analytics/days-in-inventory
func (h *KPIHandler) GetDaysInInventory(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.parseOrganizationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	snapshotDate := h.parseSnapshotDate(r)
	cacheKey := fmt.Sprintf("dii:%s:%s", orgID.String(), snapshotDate.Format("2006-01-02"))

	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	// Try to get from repository
	storedKPI, err := h.kpiRepo.GetDaysInInventoryKPI(r.Context(), orgID, snapshotDate)
	if err == nil && storedKPI != nil {
		response := dto.ToDaysInInventoryKPIResponse(storedKPI)
		h.cache.Set(cacheKey, response)
		h.respondJSON(w, http.StatusOK, response)
		return
	}

	// Calculate new KPI
	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: orgID,
		SnapshotDate:   snapshotDate,
	}

	result, err := h.diiUseCase.Execute(r.Context(), input)
	if err != nil {
		// Graceful degradation: return partial data with warning
		if h.isPartialDataError(err) {
			h.respondWithWarning(w, http.StatusOK, nil, []string{err.Error()})
			return
		}
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := dto.ToDaysInInventoryKPIResponse(result)
	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// GetImmobilizedInventory handles GET /api/v1/analytics/immobilized-inventory
func (h *KPIHandler) GetImmobilizedInventory(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.parseOrganizationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	snapshotDate := h.parseSnapshotDate(r)
	thresholdYears := h.parseThresholdYears(r)
	cacheKey := fmt.Sprintf("immobilized:%s:%s:%d", orgID.String(), snapshotDate.Format("2006-01-02"), thresholdYears)

	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	// Try to get from repository
	storedKPI, err := h.kpiRepo.GetImmobilizedInventoryKPI(r.Context(), orgID, snapshotDate, thresholdYears)
	if err == nil && storedKPI != nil {
		response := dto.ToImmobilizedInventoryKPIResponse(storedKPI)
		h.cache.Set(cacheKey, response)
		h.respondJSON(w, http.StatusOK, response)
		return
	}

	// Calculate new KPI
	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: orgID,
		SnapshotDate:   snapshotDate,
		ThresholdYears: thresholdYears,
	}

	result, err := h.immobilizedUseCase.Execute(r.Context(), input)
	if err != nil {
		if h.isPartialDataError(err) {
			h.respondWithWarning(w, http.StatusOK, nil, []string{err.Error()})
			return
		}
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := dto.ToImmobilizedInventoryKPIResponse(result)
	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// GetInventoryRotation handles GET /api/v1/analytics/inventory-rotation
func (h *KPIHandler) GetInventoryRotation(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.parseOrganizationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	snapshotDate := h.parseSnapshotDate(r)
	cacheKey := fmt.Sprintf("rotation:%s:%s", orgID.String(), snapshotDate.Format("2006-01-02"))

	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	// Try to get from repository
	storedKPI, err := h.kpiRepo.GetInventoryRotationKPI(r.Context(), orgID, snapshotDate)
	if err == nil && storedKPI != nil {
		response := dto.ToInventoryRotationKPIResponse(storedKPI)
		h.cache.Set(cacheKey, response)
		h.respondJSON(w, http.StatusOK, response)
		return
	}

	// Calculate new KPI
	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: orgID,
		SnapshotDate:   snapshotDate,
	}

	result, err := h.rotationUseCase.Execute(r.Context(), input)
	if err != nil {
		if h.isPartialDataError(err) {
			h.respondWithWarning(w, http.StatusOK, nil, []string{err.Error()})
			return
		}
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := dto.ToInventoryRotationKPIResponse(result)
	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// GetBufferAnalytics handles GET /api/v1/analytics/buffer-analytics
func (h *KPIHandler) GetBufferAnalytics(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.parseOrganizationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	// Parse optional product_id
	var productID *uuid.UUID
	if productIDStr := r.URL.Query().Get("product_id"); productIDStr != "" {
		id, err := uuid.Parse(productIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product_id")
			return
		}
		productID = &id
	}

	date := h.parseSnapshotDate(r)
	cacheKey := fmt.Sprintf("buffer:%s:%v:%s", orgID.String(), productID, date.Format("2006-01-02"))

	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	var response interface{}
	if productID != nil {
		// Get single product buffer analytics
		analytics, err := h.kpiRepo.GetBufferAnalyticsByProduct(r.Context(), orgID, *productID, date)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		response = dto.ToBufferAnalyticsResponse(analytics)
	} else {
		// Get all buffer analytics for date range
		startDate := date.AddDate(0, 0, -30)
		analytics, err := h.kpiRepo.ListBufferAnalytics(r.Context(), orgID, startDate, date)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		response = dto.ToBufferAnalyticsListResponse(analytics)
	}

	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// GetSnapshot handles GET /api/v1/analytics/snapshot
func (h *KPIHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.parseOrganizationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	snapshotDate := h.parseSnapshotDate(r)
	cacheKey := fmt.Sprintf("snapshot:%s:%s", orgID.String(), snapshotDate.Format("2006-01-02"))

	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	// Try to get from repository
	storedSnapshot, err := h.kpiRepo.GetKPISnapshot(r.Context(), orgID, snapshotDate)
	if err == nil && storedSnapshot != nil {
		response := dto.ToKPISnapshotResponse(storedSnapshot)
		h.cache.Set(cacheKey, response)
		h.respondJSON(w, http.StatusOK, response)
		return
	}

	// Build consolidated snapshot from component KPIs
	snapshot, warnings := h.buildConsolidatedSnapshot(r, orgID, snapshotDate)
	if snapshot != nil {
		response := dto.ToKPISnapshotResponse(snapshot)
		if len(warnings) > 0 {
			h.respondWithWarning(w, http.StatusOK, response, warnings)
			return
		}
		h.cache.Set(cacheKey, response)
		h.respondJSON(w, http.StatusOK, response)
		return
	}

	h.respondError(w, http.StatusNotFound, "NOT_FOUND", "no KPI data available")
}

// SyncBufferData handles POST /api/v1/analytics/sync-buffer
func (h *KPIHandler) SyncBufferData(w http.ResponseWriter, r *http.Request) {
	var req dto.SyncBufferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.OrganizationID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "organization_id is required")
		return
	}

	if h.syncBufferUseCase == nil {
		h.respondError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "buffer sync not configured")
		return
	}

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: req.OrganizationID,
		Date:           req.Date,
	}

	syncedCount, err := h.syncBufferUseCase.Execute(r.Context(), input)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Invalidate related cache
	h.cache.Clear()

	h.respondJSON(w, http.StatusOK, dto.SyncBufferResponse{
		SyncedCount: syncedCount,
		Message:     fmt.Sprintf("Successfully synced %d buffer records", syncedCount),
	})
}

// buildConsolidatedSnapshot builds a KPI snapshot from component KPIs.
func (h *KPIHandler) buildConsolidatedSnapshot(r *http.Request, orgID uuid.UUID, date time.Time) (*domain.KPISnapshot, []string) {
	var warnings []string

	rotationKPI, err := h.kpiRepo.GetInventoryRotationKPI(r.Context(), orgID, date)
	if err != nil {
		warnings = append(warnings, "rotation data unavailable")
		rotationKPI = &domain.InventoryRotationKPI{}
	}

	// Get buffer distribution (mock values for now, would be from DDMRP service)
	bufferGreen := 60.0
	bufferYellow := 30.0
	bufferRed := 10.0

	snapshot, err := domain.NewKPISnapshot(
		orgID,
		date,
		rotationKPI.RotationRatio,
		5.0,  // stockout rate - would come from execution service
		95.0, // service level - would come from execution service
		10.0, // excess inventory pct
		bufferGreen,
		bufferYellow,
		bufferRed,
		rotationKPI.AvgMonthlyStock,
	)
	if err != nil {
		return nil, warnings
	}

	return snapshot, warnings
}

// Helper methods

func (h *KPIHandler) parseOrganizationID(r *http.Request) (uuid.UUID, error) {
	orgIDStr := r.URL.Query().Get("organization_id")
	if orgIDStr == "" {
		orgIDStr = r.Header.Get("X-Organization-ID")
	}
	if orgIDStr == "" {
		return uuid.Nil, fmt.Errorf("organization_id is required")
	}
	return uuid.Parse(orgIDStr)
}

func (h *KPIHandler) parseSnapshotDate(r *http.Request) time.Time {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = r.URL.Query().Get("snapshot_date")
	}
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			return t
		}
	}
	return time.Now().UTC().Truncate(24 * time.Hour)
}

func (h *KPIHandler) parseThresholdYears(r *http.Request) int {
	thresholdStr := r.URL.Query().Get("threshold_years")
	if threshold, err := strconv.Atoi(thresholdStr); err == nil && threshold > 0 {
		return threshold
	}
	return 2 // Default threshold
}

func (h *KPIHandler) isPartialDataError(err error) bool {
	// Check if error is due to partial data availability
	return false // Implement actual check based on error types
}

func (h *KPIHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *KPIHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, dto.ErrorResponse{
		ErrorCode: code,
		Message:   message,
	})
}

func (h *KPIHandler) respondWithWarning(w http.ResponseWriter, status int, data interface{}, warnings []string) {
	h.respondJSON(w, status, dto.AnalyticsResponseWithWarnings{
		Data:     data,
		Warnings: warnings,
		Partial:  true,
	})
}
