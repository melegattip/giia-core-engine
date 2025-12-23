package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/usecases/product"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/infrastructure/entrypoints/http/dto"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	createUC *product.CreateProductUseCase
	getUC    *product.GetProductUseCase
	updateUC *product.UpdateProductUseCase
	deleteUC *product.DeleteProductUseCase
	listUC   *product.ListProductsUseCase
	searchUC *product.SearchProductsUseCase
	logger   logger.Logger
}

func NewProductHandler(
	createUC *product.CreateProductUseCase,
	getUC *product.GetProductUseCase,
	updateUC *product.UpdateProductUseCase,
	deleteUC *product.DeleteProductUseCase,
	listUC *product.ListProductsUseCase,
	searchUC *product.SearchProductsUseCase,
	logger logger.Logger,
) *ProductHandler {
	return &ProductHandler{
		createUC: createUC,
		getUC:    getUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
		listUC:   listUC,
		searchUC: searchUC,
		logger:   logger,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	ucReq := &product.CreateProductRequest{
		SKU:             req.SKU,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		UnitOfMeasure:   req.UnitOfMeasure,
		BufferProfileID: req.BufferProfileID,
	}

	createdProduct, err := h.createUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.respondError(w, err)
		return
	}

	response := dto.ToProductResponse(createdProduct)
	h.respondJSON(w, http.StatusCreated, response)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	includeSuppliers := r.URL.Query().Get("include") == "suppliers"

	fetchedProduct, err := h.getUC.Execute(r.Context(), id, includeSuppliers)
	if err != nil {
		h.respondError(w, err)
		return
	}

	response := dto.ToProductResponse(fetchedProduct)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.NewBadRequest("invalid request body"))
		return
	}

	ucReq := &product.UpdateProductRequest{
		ID:              id,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		UnitOfMeasure:   req.UnitOfMeasure,
		Status:          req.Status,
		BufferProfileID: req.BufferProfileID,
	}

	updatedProduct, err := h.updateUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.respondError(w, err)
		return
	}

	response := dto.ToProductResponse(updatedProduct)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, errors.NewBadRequest("invalid product ID"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusNoContent, nil)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")

	ucReq := &product.ListProductsRequest{
		Page:     page,
		PageSize: pageSize,
		Category: category,
		Status:   status,
	}

	result, err := h.listUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.respondError(w, err)
		return
	}

	response := dto.PaginatedProductsResponse{
		Products:   dto.ToProductListResponse(result.Products),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")

	ucReq := &product.SearchProductsRequest{
		Query:    query,
		Page:     page,
		PageSize: pageSize,
		Category: category,
		Status:   status,
	}

	result, err := h.searchUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.respondError(w, err)
		return
	}

	response := dto.PaginatedProductsResponse{
		Products:   dto.ToProductListResponse(result.Products),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *ProductHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *ProductHandler) respondError(w http.ResponseWriter, err error) {
	response := dto.ErrorResponse{
		Message: err.Error(),
	}

	status := http.StatusInternalServerError
	if customErr, ok := err.(*errors.CustomError); ok {
		response.ErrorCode = customErr.ErrorCode
		status = customErr.HTTPStatus
	}

	h.respondJSON(w, status, response)
}
