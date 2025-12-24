// Package handlers provides HTTP handlers for the Execution Service.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/purchase_order"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/dto"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

// PurchaseOrderHandler handles HTTP requests for purchase orders.
type PurchaseOrderHandler struct {
	createUC  *purchase_order.CreatePOUseCase
	receiveUC *purchase_order.ReceivePOUseCase
	cancelUC  *purchase_order.CancelPOUseCase
	poRepo    providers.PurchaseOrderRepository
}

// NewPurchaseOrderHandler creates a new purchase order handler.
func NewPurchaseOrderHandler(
	createUC *purchase_order.CreatePOUseCase,
	receiveUC *purchase_order.ReceivePOUseCase,
	cancelUC *purchase_order.CancelPOUseCase,
	poRepo providers.PurchaseOrderRepository,
) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		createUC:  createUC,
		receiveUC: receiveUC,
		cancelUC:  cancelUC,
		poRepo:    poRepo,
	}
}

// Create handles POST /api/v1/purchase-orders
func (h *PurchaseOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	var req dto.CreatePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	supplierID, err := uuid.Parse(req.SupplierID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid supplier_id")
		return
	}

	lineItems := make([]domain.POLineItem, len(req.LineItems))
	for i, item := range req.LineItems {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product_id in line item")
			return
		}
		lineItems[i] = domain.POLineItem{
			ID:        uuid.New(),
			ProductID: productID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
			LineTotal: item.Quantity * item.UnitCost,
		}
	}

	input := &purchase_order.CreatePOInput{
		OrganizationID:      orgID,
		PONumber:            req.PONumber,
		SupplierID:          supplierID,
		OrderDate:           req.OrderDate,
		ExpectedArrivalDate: req.ExpectedArrivalDate,
		LineItems:           lineItems,
		CreatedBy:           userID,
	}

	po, err := h.createUC.Execute(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto.ToPurchaseOrderResponse(po))
}

// List handles GET /api/v1/purchase-orders
func (h *PurchaseOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	status := r.URL.Query().Get("status")
	supplierID := r.URL.Query().Get("supplier_id")

	params := dto.ParseQueryParams(page, pageSize, status, supplierID, "", "")
	filters := params.ConvertFiltersToInterface()

	orders, total, err := h.poRepo.List(r.Context(), orgID, filters, params.Page, params.PageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list purchase orders")
		return
	}

	response := dto.NewPaginatedResponse(
		dto.ToPurchaseOrderListResponse(orders),
		params.Page,
		params.PageSize,
		total,
	)

	h.respondJSON(w, http.StatusOK, response)
}

// Get handles GET /api/v1/purchase-orders/{id}
func (h *PurchaseOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid purchase order ID")
		return
	}

	po, err := h.poRepo.GetByID(r.Context(), id, orgID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToPurchaseOrderResponse(po))
}

// Receive handles POST /api/v1/purchase-orders/{id}/receive
func (h *PurchaseOrderHandler) Receive(w http.ResponseWriter, r *http.Request) {
	userID, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	poID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid purchase order ID")
		return
	}

	var req dto.ReceivePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid location_id")
		return
	}

	receipts := make([]purchase_order.ReceiveLineItem, len(req.Receipts))
	for i, receipt := range req.Receipts {
		lineItemID, err := uuid.Parse(receipt.LineItemID)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid line_item_id in receipt")
			return
		}
		receipts[i] = purchase_order.ReceiveLineItem{
			LineItemID:  lineItemID,
			ReceivedQty: receipt.ReceivedQty,
		}
	}

	input := &purchase_order.ReceivePOInput{
		POID:           poID,
		OrganizationID: orgID,
		LocationID:     locationID,
		Receipts:       receipts,
		ReceivedBy:     userID,
	}

	po, err := h.receiveUC.Execute(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToPurchaseOrderResponse(po))
}

// Cancel handles POST /api/v1/purchase-orders/{id}/cancel
func (h *PurchaseOrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	poID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid purchase order ID")
		return
	}

	input := &purchase_order.CancelPOInput{
		POID:           poID,
		OrganizationID: orgID,
	}

	po, err := h.cancelUC.Execute(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToPurchaseOrderResponse(po))
}

func (h *PurchaseOrderHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *PurchaseOrderHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, dto.ErrorResponse{
		ErrorCode: code,
		Message:   message,
	})
}

func (h *PurchaseOrderHandler) handleDomainError(w http.ResponseWriter, err error) {
	if err == nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	errMsg := err.Error()

	// Check for validation errors
	if isValidationError(errMsg) {
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", errMsg)
		return
	}

	// Check for not found errors
	if isNotFoundError(errMsg) {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", errMsg)
		return
	}

	// Check for conflict errors
	if isConflictError(errMsg) {
		h.respondError(w, http.StatusConflict, "CONFLICT", errMsg)
		return
	}

	h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
}

func isValidationError(errMsg string) bool {
	// Domain validation errors typically contain these keywords
	return contains(errMsg, "is required") ||
		contains(errMsg, "invalid") ||
		contains(errMsg, "cannot be") ||
		contains(errMsg, "must be") ||
		contains(errMsg, "can only")
}

func isNotFoundError(errMsg string) bool {
	return contains(errMsg, "not found") ||
		contains(errMsg, "record not found")
}

func isConflictError(errMsg string) bool {
	return contains(errMsg, "already exists") ||
		contains(errMsg, "conflict") ||
		contains(errMsg, "duplicate")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
