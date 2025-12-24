// Package handlers provides HTTP handlers for the Execution Service.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/dto"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

// InventoryHandler handles HTTP requests for inventory operations.
type InventoryHandler struct {
	balanceRepo providers.InventoryBalanceRepository
	txnRepo     providers.InventoryTransactionRepository
}

// NewInventoryHandler creates a new inventory handler.
func NewInventoryHandler(
	balanceRepo providers.InventoryBalanceRepository,
	txnRepo providers.InventoryTransactionRepository,
) *InventoryHandler {
	return &InventoryHandler{
		balanceRepo: balanceRepo,
		txnRepo:     txnRepo,
	}
}

// GetBalances handles GET /api/v1/inventory/balances
func (h *InventoryHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	productIDStr := r.URL.Query().Get("product_id")
	locationIDStr := r.URL.Query().Get("location_id")

	var balances []*domain.InventoryBalance

	if productIDStr != "" {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product_id")
			return
		}
		balances, err = h.balanceRepo.GetByProduct(r.Context(), orgID, productID)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get inventory balances")
			return
		}
	} else if locationIDStr != "" {
		locationID, err := uuid.Parse(locationIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid location_id")
			return
		}
		balances, err = h.balanceRepo.GetByLocation(r.Context(), orgID, locationID)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get inventory balances")
			return
		}
	} else {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "product_id or location_id is required")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"balances": dto.ToInventoryBalanceListResponse(balances),
	})
}

// GetTransactions handles GET /api/v1/inventory/transactions
func (h *InventoryHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	_, orgID, err := middleware.RequireAuth(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	productIDStr := r.URL.Query().Get("product_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var productID uuid.UUID
	if productIDStr != "" {
		productID, err = uuid.Parse(productIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product_id")
			return
		}
	}

	referenceType := r.URL.Query().Get("reference_type")
	txnType := r.URL.Query().Get("type")

	filters := make(map[string]interface{})
	if referenceType != "" {
		filters["reference_type"] = referenceType
	}
	if txnType != "" {
		filters["type"] = txnType
	}

	transactions, total, err := h.txnRepo.List(r.Context(), orgID, productID, filters, page, pageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get transactions")
		return
	}

	response := dto.NewPaginatedResponse(
		dto.ToInventoryTransactionListResponse(transactions),
		page,
		pageSize,
		total,
	)

	h.respondJSON(w, http.StatusOK, response)
}

func (h *InventoryHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *InventoryHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, dto.ErrorResponse{
		ErrorCode: code,
		Message:   message,
	})
}
