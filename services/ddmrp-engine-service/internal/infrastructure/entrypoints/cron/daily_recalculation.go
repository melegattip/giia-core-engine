package cron

import (
	"context"
	"log"
	"time"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type DailyRecalculation struct {
	recalculateUseCase *buffer.RecalculateAllBuffersUseCase
	cronScheduler      *cron.Cron
	organizationIDs    []uuid.UUID
}

func NewDailyRecalculation(
	recalculateUseCase *buffer.RecalculateAllBuffersUseCase,
	organizationIDs []uuid.UUID,
) *DailyRecalculation {
	return &DailyRecalculation{
		recalculateUseCase: recalculateUseCase,
		cronScheduler:      cron.New(),
		organizationIDs:    organizationIDs,
	}
}

func (dr *DailyRecalculation) Start() {
	dr.cronScheduler.AddFunc("0 2 * * *", func() {
		log.Println("Starting daily buffer recalculation...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		for _, orgID := range dr.organizationIDs {
			if err := dr.recalculateUseCase.Execute(ctx, buffer.RecalculateAllBuffersInput{
				OrganizationID: orgID,
			}); err != nil {
				log.Printf("Error recalculating buffers for organization %s: %v", orgID, err)
			} else {
				log.Printf("Successfully recalculated buffers for organization %s", orgID)
			}
		}

		log.Println("Daily buffer recalculation completed")
	})

	dr.cronScheduler.Start()
	log.Println("Daily recalculation cron job started (runs at 2 AM daily)")
}

func (dr *DailyRecalculation) Stop() {
	dr.cronScheduler.Stop()
	log.Println("Daily recalculation cron job stopped")
}
