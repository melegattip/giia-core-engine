package server

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	catalogv1 "github.com/giia/giia-core-engine/services/catalog-service/api/proto/gen/go/catalog/v1"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	bufferProfile "github.com/giia/giia-core-engine/services/catalog-service/internal/core/usecases/buffer_profile"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/usecases/product"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/usecases/supplier"
)

type CatalogServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	createProductUC       *product.CreateProductUseCase
	getProductUC          *product.GetProductUseCase
	updateProductUC       *product.UpdateProductUseCase
	listProductsUC        *product.ListProductsUseCase
	deleteProductUC       *product.DeleteProductUseCase
	searchProductsUC      *product.SearchProductsUseCase
	createSupplierUC      *supplier.CreateSupplierUseCase
	updateSupplierUC      *supplier.UpdateSupplierUseCase
	getSupplierUC         *supplier.GetSupplierUseCase
	listSuppliersUC       *supplier.ListSuppliersUseCase
	deleteSupplierUC      *supplier.DeleteSupplierUseCase
	createBufferProfileUC *bufferProfile.CreateBufferProfileUseCase
	updateBufferProfileUC *bufferProfile.UpdateBufferProfileUseCase
	getBufferProfileUC    *bufferProfile.GetBufferProfileUseCase
	listBufferProfilesUC  *bufferProfile.ListBufferProfilesUseCase
	deleteBufferProfileUC *bufferProfile.DeleteBufferProfileUseCase
	logger                pkgLogger.Logger
}

func NewCatalogServer(
	createProductUC *product.CreateProductUseCase,
	getProductUC *product.GetProductUseCase,
	updateProductUC *product.UpdateProductUseCase,
	listProductsUC *product.ListProductsUseCase,
	deleteProductUC *product.DeleteProductUseCase,
	searchProductsUC *product.SearchProductsUseCase,
	createSupplierUC *supplier.CreateSupplierUseCase,
	updateSupplierUC *supplier.UpdateSupplierUseCase,
	getSupplierUC *supplier.GetSupplierUseCase,
	listSuppliersUC *supplier.ListSuppliersUseCase,
	deleteSupplierUC *supplier.DeleteSupplierUseCase,
	createBufferProfileUC *bufferProfile.CreateBufferProfileUseCase,
	updateBufferProfileUC *bufferProfile.UpdateBufferProfileUseCase,
	getBufferProfileUC *bufferProfile.GetBufferProfileUseCase,
	listBufferProfilesUC *bufferProfile.ListBufferProfilesUseCase,
	deleteBufferProfileUC *bufferProfile.DeleteBufferProfileUseCase,
	logger pkgLogger.Logger,
) *CatalogServer {
	return &CatalogServer{
		createProductUC:       createProductUC,
		getProductUC:          getProductUC,
		updateProductUC:       updateProductUC,
		listProductsUC:        listProductsUC,
		deleteProductUC:       deleteProductUC,
		searchProductsUC:      searchProductsUC,
		createSupplierUC:      createSupplierUC,
		updateSupplierUC:      updateSupplierUC,
		getSupplierUC:         getSupplierUC,
		listSuppliersUC:       listSuppliersUC,
		deleteSupplierUC:      deleteSupplierUC,
		createBufferProfileUC: createBufferProfileUC,
		updateBufferProfileUC: updateBufferProfileUC,
		getBufferProfileUC:    getBufferProfileUC,
		listBufferProfilesUC:  listBufferProfilesUC,
		deleteBufferProfileUC: deleteBufferProfileUC,
		logger:                logger,
	}
}

func (s *CatalogServer) CreateProduct(ctx context.Context, req *catalogv1.CreateProductRequest) (*catalogv1.CreateProductResponse, error) {
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}
	if req.Sku == "" {
		return nil, status.Error(codes.InvalidArgument, "sku is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.UnitOfMeasure == "" {
		return nil, status.Error(codes.InvalidArgument, "unit_of_measure is required")
	}

	organizationID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	var bufferProfileID *uuid.UUID
	if req.BufferProfileId != "" {
		id, err := uuid.Parse(req.BufferProfileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid buffer_profile_id format")
		}
		bufferProfileID = &id
	}

	ctx = context.WithValue(ctx, "organization_id", organizationID)

	input := &product.CreateProductRequest{
		SKU:             req.Sku,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		UnitOfMeasure:   req.UnitOfMeasure,
		BufferProfileID: bufferProfileID,
	}

	result, err := s.createProductUC.Execute(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	s.logger.Info(ctx, "Product created via gRPC", pkgLogger.Tags{
		"product_id":      result.ID.String(),
		"organization_id": req.OrganizationId,
		"sku":             req.Sku,
	})

	return &catalogv1.CreateProductResponse{
		Product: toProtoProduct(result),
	}, nil
}

func (s *CatalogServer) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id format")
	}

	organizationID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	ctx = context.WithValue(ctx, "organization_id", organizationID)

	result, err := s.getProductUC.Execute(ctx, productID, false)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &catalogv1.GetProductResponse{
		Product: toProtoProduct(result),
	}, nil
}

func (s *CatalogServer) UpdateProduct(ctx context.Context, req *catalogv1.UpdateProductRequest) (*catalogv1.UpdateProductResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id format")
	}

	organizationID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	var bufferProfileID *uuid.UUID
	if req.BufferProfileId != "" {
		id, err := uuid.Parse(req.BufferProfileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid buffer_profile_id format")
		}
		bufferProfileID = &id
	}

	ctx = context.WithValue(ctx, "organization_id", organizationID)

	input := &product.UpdateProductRequest{
		ID:              productID,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		UnitOfMeasure:   req.UnitOfMeasure,
		BufferProfileID: bufferProfileID,
		Status:          req.Status,
	}

	result, err := s.updateProductUC.Execute(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	s.logger.Info(ctx, "Product updated via gRPC", pkgLogger.Tags{
		"product_id":      req.Id,
		"organization_id": req.OrganizationId,
	})

	return &catalogv1.UpdateProductResponse{
		Product: toProtoProduct(result),
	}, nil
}

func (s *CatalogServer) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	organizationID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx = context.WithValue(ctx, "organization_id", organizationID)

	input := &product.ListProductsRequest{
		Page:     int(page),
		PageSize: int(pageSize),
		Status:   req.Status,
		Category: req.Category,
	}

	result, err := s.listProductsUC.Execute(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	products := make([]*catalogv1.Product, len(result.Products))
	for i, p := range result.Products {
		products[i] = toProtoProduct(p)
	}

	return &catalogv1.ListProductsResponse{
		Products: products,
		Total:    int32(result.TotalCount),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *CatalogServer) DeleteProduct(ctx context.Context, req *catalogv1.DeleteProductRequest) (*catalogv1.DeleteProductResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id format")
	}

	organizationID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	ctx = context.WithValue(ctx, "organization_id", organizationID)

	err = s.deleteProductUC.Execute(ctx, productID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	s.logger.Info(ctx, "Product deleted via gRPC", pkgLogger.Tags{
		"product_id":      req.Id,
		"organization_id": req.OrganizationId,
	})

	return &catalogv1.DeleteProductResponse{
		Success: true,
	}, nil
}

func (s *CatalogServer) SearchProducts(ctx context.Context, req *catalogv1.SearchProductsRequest) (*catalogv1.SearchProductsResponse, error) {
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	organizationID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx = context.WithValue(ctx, "organization_id", organizationID)

	input := &product.SearchProductsRequest{
		Query:    req.Query,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	result, err := s.searchProductsUC.Execute(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	products := make([]*catalogv1.Product, len(result.Products))
	for i, p := range result.Products {
		products[i] = toProtoProduct(p)
	}

	return &catalogv1.SearchProductsResponse{
		Products: products,
		Total:    int32(result.TotalCount),
	}, nil
}

func (s *CatalogServer) GetSupplier(ctx context.Context, req *catalogv1.GetSupplierRequest) (*catalogv1.GetSupplierResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	supplierID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid supplier ID format")
	}

	result, err := s.getSupplierUC.Execute(ctx, supplierID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &catalogv1.GetSupplierResponse{
		Supplier: toProtoSupplier(result),
	}, nil
}

func (s *CatalogServer) ListSuppliers(ctx context.Context, req *catalogv1.ListSuppliersRequest) (*catalogv1.ListSuppliersResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := s.listSuppliersUC.Execute(ctx, &supplier.ListSuppliersRequest{
		Status:   req.Status,
		Page:     int(page),
		PageSize: int(pageSize),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	suppliers := make([]*catalogv1.Supplier, len(result.Suppliers))
	for i, sup := range result.Suppliers {
		suppliers[i] = toProtoSupplier(sup)
	}

	return &catalogv1.ListSuppliersResponse{
		Suppliers: suppliers,
		Total:     int32(result.Total),
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

func (s *CatalogServer) GetBufferProfile(ctx context.Context, req *catalogv1.GetBufferProfileRequest) (*catalogv1.GetBufferProfileResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	profileID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid buffer profile ID format")
	}

	result, err := s.getBufferProfileUC.Execute(ctx, profileID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &catalogv1.GetBufferProfileResponse{
		BufferProfile: toProtoBufferProfile(result),
	}, nil
}

func (s *CatalogServer) ListBufferProfiles(ctx context.Context, req *catalogv1.ListBufferProfilesRequest) (*catalogv1.ListBufferProfilesResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := s.listBufferProfilesUC.Execute(ctx, &bufferProfile.ListBufferProfilesRequest{
		Page:     int(page),
		PageSize: int(pageSize),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	profiles := make([]*catalogv1.BufferProfile, len(result.BufferProfiles))
	for i, prof := range result.BufferProfiles {
		profiles[i] = toProtoBufferProfile(prof)
	}

	return &catalogv1.ListBufferProfilesResponse{
		BufferProfiles: profiles,
		Total:          int32(result.Total),
		Page:           page,
		PageSize:       pageSize,
	}, nil
}

func (s *CatalogServer) CreateSupplier(ctx context.Context, req *catalogv1.CreateSupplierRequest) (*catalogv1.CreateSupplierResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateSupplier not implemented")
}

func (s *CatalogServer) UpdateSupplier(ctx context.Context, req *catalogv1.UpdateSupplierRequest) (*catalogv1.UpdateSupplierResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateSupplier not implemented")
}

func (s *CatalogServer) DeleteSupplier(ctx context.Context, req *catalogv1.DeleteSupplierRequest) (*catalogv1.DeleteSupplierResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteSupplier not implemented")
}

func (s *CatalogServer) CreateBufferProfile(ctx context.Context, req *catalogv1.CreateBufferProfileRequest) (*catalogv1.CreateBufferProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateBufferProfile not implemented")
}

func (s *CatalogServer) UpdateBufferProfile(ctx context.Context, req *catalogv1.UpdateBufferProfileRequest) (*catalogv1.UpdateBufferProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateBufferProfile not implemented")
}

func (s *CatalogServer) DeleteBufferProfile(ctx context.Context, req *catalogv1.DeleteBufferProfileRequest) (*catalogv1.DeleteBufferProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteBufferProfile not implemented")
}

func (s *CatalogServer) AssociateSupplier(ctx context.Context, req *catalogv1.AssociateSupplierRequest) (*catalogv1.AssociateSupplierResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method AssociateSupplier not implemented")
}

func (s *CatalogServer) GetProductSuppliers(ctx context.Context, req *catalogv1.GetProductSuppliersRequest) (*catalogv1.GetProductSuppliersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetProductSuppliers not implemented")
}

func (s *CatalogServer) RemoveSupplierAssociation(ctx context.Context, req *catalogv1.RemoveSupplierAssociationRequest) (*catalogv1.RemoveSupplierAssociationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RemoveSupplierAssociation not implemented")
}

func toProtoProduct(p *domain.Product) *catalogv1.Product {
	proto := &catalogv1.Product{
		Id:             p.ID.String(),
		OrganizationId: p.OrganizationID.String(),
		Sku:            p.SKU,
		Name:           p.Name,
		Description:    p.Description,
		Category:       p.Category,
		UnitOfMeasure:  p.UnitOfMeasure,
		Status:         string(p.Status),
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
	}

	if p.BufferProfileID != nil {
		proto.BufferProfileId = p.BufferProfileID.String()
	}

	return proto
}

func toProtoSupplier(s *domain.Supplier) *catalogv1.Supplier {
	return &catalogv1.Supplier{
		Id:             s.ID.String(),
		OrganizationId: s.OrganizationID.String(),
		Code:           s.Code,
		Name:           s.Name,
		Status:         string(s.Status),
		CreatedAt:      timestamppb.New(s.CreatedAt),
		UpdatedAt:      timestamppb.New(s.UpdatedAt),
	}
}

func toProtoBufferProfile(bp *domain.BufferProfile) *catalogv1.BufferProfile {
	return &catalogv1.BufferProfile{
		Id:                bp.ID.String(),
		OrganizationId:    bp.OrganizationID.String(),
		Name:              bp.Name,
		Description:       bp.Description,
		LeadTimeFactor:    bp.LeadTimeFactor,
		VariabilityFactor: bp.VariabilityFactor,
		CreatedAt:         timestamppb.New(bp.CreatedAt),
		UpdatedAt:         timestamppb.New(bp.UpdatedAt),
	}
}

func mapDomainError(err error) error {
	var customErr *pkgErrors.CustomError
	if errors.As(err, &customErr) {
		switch customErr.ErrorCode {
		case "NOT_FOUND":
			return status.Error(codes.NotFound, err.Error())
		case "BAD_REQUEST":
			return status.Error(codes.InvalidArgument, err.Error())
		case "UNAUTHORIZED":
			return status.Error(codes.Unauthenticated, err.Error())
		case "FORBIDDEN":
			return status.Error(codes.PermissionDenied, err.Error())
		case "CONFLICT":
			return status.Error(codes.AlreadyExists, err.Error())
		}
	}

	return status.Error(codes.Internal, "internal server error")
}
