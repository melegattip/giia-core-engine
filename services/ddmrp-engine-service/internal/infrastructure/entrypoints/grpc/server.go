package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/api/proto/ddmrp/v1"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/demand_adjustment"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/nfp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CalculateBufferUseCase interface for buffer calculation
type CalculateBufferUseCase interface {
	Execute(ctx context.Context, input buffer.CalculateBufferInput) (*domain.Buffer, error)
}

// GetBufferUseCase interface for getting a buffer
type GetBufferUseCase interface {
	Execute(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error)
}

// ListBuffersUseCase interface for listing buffers
type ListBuffersUseCase interface {
	Execute(ctx context.Context, input buffer.ListBuffersInput) ([]domain.Buffer, error)
}

// CreateFADUseCase interface for creating FAD
type CreateFADUseCase interface {
	Execute(ctx context.Context, input demand_adjustment.CreateFADInput) (*domain.DemandAdjustment, error)
}

// UpdateFADUseCase interface for updating FAD
type UpdateFADUseCase interface {
	Execute(ctx context.Context, input demand_adjustment.UpdateFADInput) (*domain.DemandAdjustment, error)
}

// DeleteFADUseCase interface for deleting FAD
type DeleteFADUseCase interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

// ListFADsUseCase interface for listing FADs
type ListFADsUseCase interface {
	ExecuteByProduct(ctx context.Context, input demand_adjustment.ListFADsByProductInput) ([]domain.DemandAdjustment, error)
	ExecuteByOrganization(ctx context.Context, input demand_adjustment.ListFADsByOrganizationInput) ([]domain.DemandAdjustment, error)
}

// UpdateNFPUseCase interface for updating NFP
type UpdateNFPUseCase interface {
	Execute(ctx context.Context, input nfp.UpdateNFPInput) (*domain.Buffer, error)
}

// CheckReplenishmentUseCase interface for checking replenishment
type CheckReplenishmentUseCase interface {
	Execute(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error)
}

// DDMRPServiceServer implements the gRPC DDMRPService
type DDMRPServiceServer struct {
	pb.UnimplementedDDMRPServiceServer
	calculateBufferUC    CalculateBufferUseCase
	getBufferUC          GetBufferUseCase
	listBuffersUC        ListBuffersUseCase
	createFADUC          CreateFADUseCase
	updateFADUC          UpdateFADUseCase
	deleteFADUC          DeleteFADUseCase
	listFADsUC           ListFADsUseCase
	updateNFPUC          UpdateNFPUseCase
	checkReplenishmentUC CheckReplenishmentUseCase
}

// NewDDMRPServiceServer creates a new DDMRPServiceServer
func NewDDMRPServiceServer(
	calculateBufferUC CalculateBufferUseCase,
	getBufferUC GetBufferUseCase,
	listBuffersUC ListBuffersUseCase,
	createFADUC CreateFADUseCase,
	updateFADUC UpdateFADUseCase,
	deleteFADUC DeleteFADUseCase,
	listFADsUC ListFADsUseCase,
	updateNFPUC UpdateNFPUseCase,
	checkReplenishmentUC CheckReplenishmentUseCase,
) *DDMRPServiceServer {
	return &DDMRPServiceServer{
		calculateBufferUC:    calculateBufferUC,
		getBufferUC:          getBufferUC,
		listBuffersUC:        listBuffersUC,
		createFADUC:          createFADUC,
		updateFADUC:          updateFADUC,
		deleteFADUC:          deleteFADUC,
		listFADsUC:           listFADsUC,
		updateNFPUC:          updateNFPUC,
		checkReplenishmentUC: checkReplenishmentUC,
	}
}

// CalculateBuffer implements DDMRPServiceServer
func (s *DDMRPServiceServer) CalculateBuffer(ctx context.Context, req *pb.CalculateBufferRequest) (*pb.CalculateBufferResponse, error) {
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, domain.NewValidationError("invalid product_id: " + err.Error())
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	buf, err := s.calculateBufferUC.Execute(ctx, buffer.CalculateBufferInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CalculateBufferResponse{
		Buffer: domainBufferToProto(buf),
	}, nil
}

// GetBuffer implements DDMRPServiceServer
func (s *DDMRPServiceServer) GetBuffer(ctx context.Context, req *pb.GetBufferRequest) (*pb.GetBufferResponse, error) {
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, domain.NewValidationError("invalid product_id: " + err.Error())
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	buf, err := s.getBufferUC.Execute(ctx, productID, orgID)
	if err != nil {
		return nil, err
	}

	return &pb.GetBufferResponse{
		Buffer: domainBufferToProto(buf),
	}, nil
}

// ListBuffers implements DDMRPServiceServer
func (s *DDMRPServiceServer) ListBuffers(ctx context.Context, req *pb.ListBuffersRequest) (*pb.ListBuffersResponse, error) {
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	input := buffer.ListBuffersInput{
		OrganizationID: orgID,
		Zone:           domain.ZoneType(req.Zone),
		AlertLevel:     domain.AlertLevel(req.AlertLevel),
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
	}

	buffers, err := s.listBuffersUC.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	protoBuffers := make([]*pb.Buffer, len(buffers))
	for i, buf := range buffers {
		protoBuffers[i] = domainBufferToProto(&buf)
	}

	return &pb.ListBuffersResponse{
		Buffers: protoBuffers,
	}, nil
}

// CreateFAD implements DDMRPServiceServer
func (s *DDMRPServiceServer) CreateFAD(ctx context.Context, req *pb.CreateFADRequest) (*pb.CreateFADResponse, error) {
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, domain.NewValidationError("invalid product_id: " + err.Error())
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, domain.NewValidationError("invalid created_by: " + err.Error())
	}

	fad, err := s.createFADUC.Execute(ctx, demand_adjustment.CreateFADInput{
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      req.StartDate.AsTime(),
		EndDate:        req.EndDate.AsTime(),
		AdjustmentType: domain.DemandAdjustmentType(req.AdjustmentType),
		Factor:         req.Factor,
		Reason:         req.Reason,
		CreatedBy:      createdBy,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateFADResponse{
		DemandAdjustment: domainFADToProto(fad),
	}, nil
}

// UpdateFAD implements DDMRPServiceServer
func (s *DDMRPServiceServer) UpdateFAD(ctx context.Context, req *pb.UpdateFADRequest) (*pb.UpdateFADResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, domain.NewValidationError("invalid id: " + err.Error())
	}

	fad, err := s.updateFADUC.Execute(ctx, demand_adjustment.UpdateFADInput{
		ID:             id,
		StartDate:      req.StartDate.AsTime(),
		EndDate:        req.EndDate.AsTime(),
		AdjustmentType: domain.DemandAdjustmentType(req.AdjustmentType),
		Factor:         req.Factor,
		Reason:         req.Reason,
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateFADResponse{
		DemandAdjustment: domainFADToProto(fad),
	}, nil
}

// DeleteFAD implements DDMRPServiceServer
func (s *DDMRPServiceServer) DeleteFAD(ctx context.Context, req *pb.DeleteFADRequest) (*pb.DeleteFADResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, domain.NewValidationError("invalid id: " + err.Error())
	}

	err = s.deleteFADUC.Execute(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteFADResponse{
		Success: true,
	}, nil
}

// ListFADs implements DDMRPServiceServer
func (s *DDMRPServiceServer) ListFADs(ctx context.Context, req *pb.ListFADsRequest) (*pb.ListFADsResponse, error) {
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, domain.NewValidationError("invalid product_id: " + err.Error())
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	fads, err := s.listFADsUC.ExecuteByProduct(ctx, demand_adjustment.ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, err
	}

	protoFADs := make([]*pb.DemandAdjustment, len(fads))
	for i, fad := range fads {
		protoFADs[i] = domainFADToProto(&fad)
	}

	return &pb.ListFADsResponse{
		DemandAdjustments: protoFADs,
	}, nil
}

// UpdateNFP implements DDMRPServiceServer
func (s *DDMRPServiceServer) UpdateNFP(ctx context.Context, req *pb.UpdateNFPRequest) (*pb.UpdateNFPResponse, error) {
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, domain.NewValidationError("invalid product_id: " + err.Error())
	}

	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	buf, err := s.updateNFPUC.Execute(ctx, nfp.UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          req.OnHand,
		OnOrder:         req.OnOrder,
		QualifiedDemand: req.QualifiedDemand,
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateNFPResponse{
		Buffer: domainBufferToProto(buf),
	}, nil
}

// CheckReplenishment implements DDMRPServiceServer
func (s *DDMRPServiceServer) CheckReplenishment(ctx context.Context, req *pb.CheckReplenishmentRequest) (*pb.CheckReplenishmentResponse, error) {
	orgID, err := uuid.Parse(req.OrganizationId)
	if err != nil {
		return nil, domain.NewValidationError("invalid organization_id: " + err.Error())
	}

	buffers, err := s.checkReplenishmentUC.Execute(ctx, orgID)
	if err != nil {
		return nil, err
	}

	protoBuffers := make([]*pb.Buffer, len(buffers))
	for i, buf := range buffers {
		protoBuffers[i] = domainBufferToProto(&buf)
	}

	return &pb.CheckReplenishmentResponse{
		Buffers: protoBuffers,
	}, nil
}

// domainBufferToProto converts a domain Buffer to a proto Buffer
func domainBufferToProto(buf *domain.Buffer) *pb.Buffer {
	return &pb.Buffer{
		Id:                 buf.ID.String(),
		ProductId:          buf.ProductID.String(),
		OrganizationId:     buf.OrganizationID.String(),
		BufferProfileId:    buf.BufferProfileID.String(),
		Cpd:                buf.CPD,
		Ltd:                int32(buf.LTD),
		RedBase:            buf.RedBase,
		RedSafe:            buf.RedSafe,
		RedZone:            buf.RedZone,
		YellowZone:         buf.YellowZone,
		GreenZone:          buf.GreenZone,
		TopOfRed:           buf.TopOfRed,
		TopOfYellow:        buf.TopOfYellow,
		TopOfGreen:         buf.TopOfGreen,
		OnHand:             buf.OnHand,
		OnOrder:            buf.OnOrder,
		QualifiedDemand:    buf.QualifiedDemand,
		NetFlowPosition:    buf.NetFlowPosition,
		BufferPenetration:  buf.BufferPenetration,
		Zone:               string(buf.Zone),
		AlertLevel:         string(buf.AlertLevel),
		LastRecalculatedAt: timestamppb.New(buf.LastRecalculatedAt),
		CreatedAt:          timestamppb.New(buf.CreatedAt),
		UpdatedAt:          timestamppb.New(buf.UpdatedAt),
	}
}

// domainFADToProto converts a domain DemandAdjustment to a proto DemandAdjustment
func domainFADToProto(fad *domain.DemandAdjustment) *pb.DemandAdjustment {
	return &pb.DemandAdjustment{
		Id:             fad.ID.String(),
		ProductId:      fad.ProductID.String(),
		OrganizationId: fad.OrganizationID.String(),
		StartDate:      timestamppb.New(fad.StartDate),
		EndDate:        timestamppb.New(fad.EndDate),
		AdjustmentType: string(fad.AdjustmentType),
		Factor:         fad.Factor,
		Reason:         fad.Reason,
		CreatedAt:      timestamppb.New(fad.CreatedAt),
		CreatedBy:      fad.CreatedBy.String(),
	}
}
