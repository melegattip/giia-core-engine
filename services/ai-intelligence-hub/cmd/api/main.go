package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/melegattip/giia-core-engine/pkg/config"
	"github.com/melegattip/giia-core-engine/pkg/database"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/usecases/analysis"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/usecases/event_processing"
	aiAdapter "github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/infrastructure/adapters/ai"
	eventsAdapter "github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/infrastructure/adapters/events"
	ragAdapter "github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/infrastructure/adapters/rag"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/infrastructure/repositories"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New("AI_HUB")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logLevel := cfg.GetString("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	appLogger := logger.New("ai-intelligence-hub", logLevel)
	appLogger.Info(ctx, "Starting AI Intelligence Hub service", logger.Tags{
		"version": "1.0.0",
	})

	dbConfig := &database.Config{
		Host:         cfg.GetString("DB_HOST"),
		Port:         cfg.GetInt("DB_PORT"),
		User:         cfg.GetString("DB_USER"),
		Password:     cfg.GetString("DB_PASSWORD"),
		DatabaseName: cfg.GetString("DB_NAME"),
		SSLMode:      cfg.GetString("DB_SSLMODE"),
	}
	if dbConfig.Port == 0 {
		dbConfig.Port = 5432
	}

	db := database.New()
	gormDB, err := db.Connect(ctx, dbConfig)
	if err != nil {
		appLogger.Fatal(ctx, err, "Failed to connect to database", nil)
	}
	sqlDB, _ := gormDB.DB()
	defer sqlDB.Close()

	appLogger.Info(ctx, "Database connection established", nil)

	natsServers := cfg.GetString("NATS_SERVERS")
	if natsServers == "" {
		natsServers = "nats://localhost:4222"
	}
	natsConn, err := events.ConnectWithDefaults(natsServers)
	if err != nil {
		appLogger.Fatal(ctx, err, "Failed to connect to NATS", nil)
	}
	defer natsConn.Close()

	appLogger.Info(ctx, "NATS connection established", nil)

	subscriber, err := events.NewSubscriber(natsConn)
	if err != nil {
		appLogger.Fatal(ctx, err, "Failed to create NATS subscriber", nil)
	}

	notificationRepo := repositories.NewNotificationRepository(sqlDB)

	claudeAPIKey := cfg.GetString("CLAUDE_API_KEY")
	claudeModel := cfg.GetString("CLAUDE_MODEL")
	if claudeModel == "" {
		claudeModel = "claude-3-sonnet-20240229"
	}
	aiClient := aiAdapter.NewClaudeClient(claudeAPIKey, claudeModel, appLogger)

	knowledgePath := "./knowledge_base"
	ragRetriever := ragAdapter.NewSimpleKnowledgeRetriever(knowledgePath, appLogger)
	if err := ragRetriever.Initialize(ctx); err != nil {
		appLogger.Warn(ctx, "Failed to initialize knowledge base", logger.Tags{
			"error": err.Error(),
		})
	}

	stockoutAnalyzer := analysis.NewAnalyzeStockoutRiskUseCase(
		aiClient,
		ragRetriever,
		notificationRepo,
		appLogger,
	)

	bufferHandler := event_processing.NewBufferEventHandler(stockoutAnalyzer, appLogger)
	executionHandler := event_processing.NewExecutionEventHandler(notificationRepo, appLogger)
	userHandler := event_processing.NewUserEventHandler(notificationRepo, appLogger)

	eventSubscriber := eventsAdapter.NewNATSEventSubscriber(
		subscriber,
		bufferHandler,
		executionHandler,
		userHandler,
		appLogger,
	)

	if err := eventSubscriber.Start(ctx); err != nil {
		appLogger.Fatal(ctx, err, "Failed to start event subscriber", nil)
	}

	appLogger.Info(ctx, "AI Intelligence Hub service started successfully", logger.Tags{
		"service": "ai-intelligence-hub",
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	appLogger.Info(ctx, "Shutting down gracefully...", nil)

	if err := eventSubscriber.Stop(); err != nil {
		appLogger.Error(ctx, err, "Error stopping event subscriber", nil)
	}
}
