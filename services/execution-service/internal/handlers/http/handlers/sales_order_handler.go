// Package handlers provides HTTP handlers for the Execution Service.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/sales_order"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/dto"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

// SalesOrderHandler handles HTTP requests for sales orders.
type SalesOrderHandler struct {
	createUC   *sales_order.CreateSOUseCase
	shipUC     *sales_order.IssueDeliveryNoteUseCase
	soRepo     providers.SalesOrderRepository
	cancelFunc func(*domain.SalesOrder) error
}

// NewSalesOrderHandler creates a new sales order handler.
func NewSalesOrderHandler(
	createUC *sales_order.CreateSOUseCase,
	shipUC *sales_order.IssueDeliveryNoteUseCase,
	soRepo providers.SalesOrderRepository,
) *SalesOrderHandler {
	return &SalesOrderHandler{
		createUC: createUC,
		shipUC:   shipUC,
		soRepo:   soRepo,
	}
}

// Create handles POST /api/v1/sales-orders
func (h *SalesOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	var req dto.CreateSalesOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid customer_id")
		return
	}

	lineItems := make([]domain.SOLineItem, len(req.LineItems))
	for i, item := range req.LineItems {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product_id in line item")
			return
		}
		lineItems[i] = domain.SOLineItem{
			ID:        uuid.New(),
			ProductID: productID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			LineTotal: item.Quantity * item.UnitPrice,
		}
	}

	input := &sales_order.CreateSOInput{
		OrganizationID: orgID,
		SONumber:       req.SONumber,
		CustomerID:     customerID,
		OrderDate:      req.OrderDate,
		DueDate:        req.DueDate,
		LineItems:      lineItems,
	}

	so, err := h.createUC.Execute(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto.ToSalesOrderResponse(so))
}

// List handles GET /api/v1/sales-orders
func (h *SalesOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	status := r.URL.Query().Get("status")
	customerID := r.URL.Query().Get("customer_id")

	params := dto.ParseQueryParams(page, pageSize, status, "", customerID, "")
	filters := params.ConvertFiltersToInterface()

	orders, total, err := h.soRepo.List(r.Context(), orgID, filters, params.Page, params.PageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list sales orders")
		return
	}

	response := dto.NewPaginatedResponse(
		dto.ToSalesOrderListResponse(orders),
		params.Page,
		params.PageSize,
		total,
	)

	h.respondJSON(w, http.StatusOK, response)
}

// Get handles GET /api/v1/sales-orders/{id}
func (h *SalesOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid sales order ID")
		return
	}

	so, err := h.soRepo.GetByID(r.Context(), id, orgID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToSalesOrderResponse(so))
}

// Ship handles POST /api/v1/sales-orders/{id}/ship
func (h *SalesOrderHandler) Ship(w http.ResponseWriter, r *http.Request) {
	userID, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	soID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid sales order ID")
		return
	}

	var req dto.ShipSalesOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid location_id")
		return
	}

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               soID,
		OrganizationID:     orgID,
		LocationID:         locationID,
		DeliveryNoteNumber: req.DeliveryNoteNumber,
		IssuedBy:           userID,
	}

	so, err := h.shipUC.Execute(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToSalesOrderResponse(so))
}

// Cancel handles POST /api/v1/sales-orders/{id}/cancel
func (h *SalesOrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	soID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid sales order ID")
		return
	}

	so, err := h.soRepo.GetByID(r.Context(), soID, orgID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	if err := so.Cancel(); err != nil {
		h.handleDomainError(w, err)
		return
	}

	so.UpdatedAt = time.Now()

	if err := h.soRepo.Update(r.Context(), so); err != nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to cancel sales order")
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToSalesOrderResponse(so))
}

func (h *SalesOrderHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *SalesOrderHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, dto.ErrorResponse{
		ErrorCode: code,
		Message:   message,
	})
}

func (h *SalesOrderHandler) handleDomainError(w http.ResponseWriter, err error) {
	if err == nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	errMsg := err.Error()

	// Check for validation errors
	if isSOValidationError(errMsg) {
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", errMsg)
		return
	}

	// Check for not found errors
	if isSONotFoundError(errMsg) {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", errMsg)
		return
	}

	// Check for conflict errors
	if isSOConflictError(errMsg) {
		h.respondError(w, http.StatusConflict, "CONFLICT", errMsg)
		return
	}

	h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
}

func isSOValidationError(errMsg string) bool {
	return containsStr(errMsg, "is required") ||
		containsStr(errMsg, "invalid") ||
		containsStr(errMsg, "cannot be") ||
		containsStr(errMsg, "must be") ||
		containsStr(errMsg, "can only")
}

func isSONotFoundError(errMsg string) bool {
	return containsStr(errMsg, "not found") ||
		containsStr(errMsg, "record not found")
}

func isSOConflictError(errMsg string) bool {
	return containsStr(errMsg, "already exists") ||
		containsStr(errMsg, "conflict") ||
		containsStr(errMsg, "duplicate")
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
