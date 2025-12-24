package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/google/uuid"
	pb "github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/api/proto/ddmrp/v1"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/demand_adjustment"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/nfp"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/adapters/catalog"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/adapters/events"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/config"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/entrypoints/cron"
	grpchandlers "github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/entrypoints/grpc"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/repositories"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&domain.ADUCalculation{},
		&domain.Buffer{},
		&domain.DemandAdjustment{},
		&domain.BufferAdjustment{},
		&domain.BufferHistory{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("✅ Database connection established and tables migrated")

	// Initialize repositories
	bufferRepo := repositories.NewBufferRepository(db)
	demandAdjRepo := repositories.NewDemandAdjustmentRepository(db)
	bufferAdjRepo := repositories.NewBufferAdjustmentRepository(db)
	bufferHistoryRepo := repositories.NewBufferHistoryRepository(db)
	aduRepo := repositories.NewADURepository(db)

	// Initialize external clients
	catalogClient, err := catalog.NewCatalogGRPCClient(cfg.Catalog.GRPCURL)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to catalog service: %v", err)
	}

	eventPublisher := events.NewNATSPublisher()

	// Initialize Buffer use cases
	calculateBufferUC := buffer.NewCalculateBufferUseCase(
		bufferRepo,
		demandAdjRepo,
		bufferAdjRepo,
		bufferHistoryRepo,
		catalogClient,
		aduRepo,
		eventPublisher,
	)
	getBufferUC := buffer.NewGetBufferUseCase(bufferRepo)
	listBuffersUC := buffer.NewListBuffersUseCase(bufferRepo)
	recalculateAllUC := buffer.NewRecalculateAllBuffersUseCase(calculateBufferUC, listBuffersUC)

	// Initialize FAD (Demand Adjustment) use cases
	createFADUC := demand_adjustment.NewCreateFADUseCase(demandAdjRepo, eventPublisher)
	updateFADUC := demand_adjustment.NewUpdateFADUseCase(demandAdjRepo, eventPublisher)
	deleteFADUC := demand_adjustment.NewDeleteFADUseCase(demandAdjRepo, eventPublisher)
	listFADsUC := demand_adjustment.NewListFADsUseCase(demandAdjRepo)

	// Initialize NFP use cases
	updateNFPUC := nfp.NewUpdateNFPUseCase(bufferRepo, eventPublisher)
	checkReplenishmentUC := nfp.NewCheckReplenishmentUseCase(bufferRepo)

	// Initialize gRPC service server
	ddmrpService := grpchandlers.NewDDMRPServiceServer(
		calculateBufferUC,
		getBufferUC,
		listBuffersUC,
		createFADUC,
		updateFADUC,
		deleteFADUC,
		listFADsUC,
		updateNFPUC,
		checkReplenishmentUC,
	)

	log.Println("✅ DDMRP gRPC service initialized")

	// Start cron job if enabled
	if cfg.Cron.Enabled {
		cronJob := cron.NewDailyRecalculation(recalculateAllUC, []uuid.UUID{})
		cronJob.Start()
		defer cronJob.Stop()
		log.Println("✅ Daily recalculation cron job started")
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer()

	// Register DDMRP gRPC service
	pb.RegisterDDMRPServiceServer(grpcServer, ddmrpService)

	// Enable gRPC reflection for grpcurl and debugging
	reflection.Register(grpcServer)

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		log.Printf("🚀 gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// HTTP health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"ddmrp-engine-service"}`))
	})

	fmt.Printf("🚀 DDMRP Engine Service starting on HTTP port %s...\n", cfg.Server.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.Server.HTTPPort, nil); err != nil {
		log.Fatal(err)
	}
}
