package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/entrypoints/http/dto"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewHealthHandler(db *gorm.DB, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := dto.HealthResponse{
		Status:  "ok",
		Service: "catalog-service",
		Checks:  make(map[string]string),
	}

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		response.Status = "degraded"
		response.Checks["database"] = "unhealthy"
	} else {
		response.Checks["database"] = "healthy"
	}

	status := http.StatusOK
	if response.Status != "ok" {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	h.respondJSON(w, status, response)
}

func (h *HealthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
