package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/demand_adjustment"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/nfp"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/adapters/catalog"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/adapters/events"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/config"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/entrypoints/cron"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/infrastructure/repositories"
	"github.com/google/uuid"
	"google.golang.org/grpc"
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

	bufferRepo := repositories.NewBufferRepository(db)
	demandAdjRepo := repositories.NewDemandAdjustmentRepository(db)
	bufferAdjRepo := repositories.NewBufferAdjustmentRepository(db)
	bufferHistoryRepo := repositories.NewBufferHistoryRepository(db)
	aduRepo := repositories.NewADURepository(db)

	catalogClient, err := catalog.NewCatalogGRPCClient(cfg.Catalog.GRPCURL)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to catalog service: %v", err)
	}

	eventPublisher := events.NewNATSPublisher()

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

	createFADUC := demand_adjustment.NewCreateFADUseCase(demandAdjRepo, eventPublisher)
	updateFADUC := demand_adjustment.NewUpdateFADUseCase(demandAdjRepo, eventPublisher)
	deleteFADUC := demand_adjustment.NewDeleteFADUseCase(demandAdjRepo, eventPublisher)
	listFADsUC := demand_adjustment.NewListFADsUseCase(demandAdjRepo)

	updateNFPUC := nfp.NewUpdateNFPUseCase(bufferRepo, eventPublisher)
	checkReplenishmentUC := nfp.NewCheckReplenishmentUseCase(bufferRepo)

	_ = createFADUC
	_ = updateFADUC
	_ = deleteFADUC
	_ = listFADsUC
	_ = updateNFPUC
	_ = checkReplenishmentUC
	_ = getBufferUC

	if cfg.Cron.Enabled {
		cronJob := cron.NewDailyRecalculation(recalculateAllUC, []uuid.UUID{})
		cronJob.Start()
		defer cronJob.Stop()
		log.Println("✅ Daily recalculation cron job started")
	}

	grpcServer := grpc.NewServer()

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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"ddmrp-engine-service"}`))
	})

	fmt.Printf("🚀 DDMRP Engine Service starting on HTTP port %s...\n", cfg.Server.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.Server.HTTPPort, nil); err != nil {
		log.Fatal(err)
	}
}
