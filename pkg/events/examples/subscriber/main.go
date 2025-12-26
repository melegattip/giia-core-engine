package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/events"
)

type UserEventProcessor struct {
	subscriber events.Subscriber
}

func NewUserEventProcessor(subscriber events.Subscriber) *UserEventProcessor {
	return &UserEventProcessor{
		subscriber: subscriber,
	}
}

func (p *UserEventProcessor) HandleUserCreated(ctx context.Context, event *events.Event) error {
	userID := event.Data["user_id"].(float64)
	email := event.Data["email"].(string)
	role := event.Data["role"].(string)

	log.Printf("Processing user.created event: user_id=%d, email=%s, role=%s",
		int64(userID), email, role)

	return nil
}

func (p *UserEventProcessor) HandleUserRoleUpdated(ctx context.Context, event *events.Event) error {
	userID := event.Data["user_id"].(float64)
	oldRole := event.Data["old_role"].(string)
	newRole := event.Data["new_role"].(string)

	log.Printf("Processing user.role.updated event: user_id=%d, old_role=%s, new_role=%s",
		int64(userID), oldRole, newRole)

	return nil
}

func (p *UserEventProcessor) HandleUserDeleted(ctx context.Context, event *events.Event) error {
	userID := event.Data["user_id"].(float64)
	reason := event.Data["reason"].(string)

	log.Printf("Processing user.deleted event: user_id=%d, reason=%s",
		int64(userID), reason)

	return nil
}

func (p *UserEventProcessor) RouteEvent(ctx context.Context, event *events.Event) error {
	log.Printf("Received event: type=%s, source=%s, org=%s, timestamp=%s",
		event.Type, event.Source, event.OrganizationID, event.Timestamp.Format(time.RFC3339))

	switch event.Type {
	case "user.created":
		return p.HandleUserCreated(ctx, event)
	case "user.role.updated":
		return p.HandleUserRoleUpdated(ctx, event)
	case "user.deleted":
		return p.HandleUserDeleted(ctx, event)
	default:
		log.Printf("Unknown event type: %s", event.Type)
		return nil
	}
}

func (p *UserEventProcessor) Start(ctx context.Context) error {
	config := &events.SubscriberConfig{
		MaxDeliver: 5,
		AckWait:    30 * time.Second,
	}

	return p.subscriber.SubscribeDurableWithConfig(
		ctx,
		"auth.user.*",
		"catalog-service-user-consumer",
		config,
		p.RouteEvent,
	)
}

func (p *UserEventProcessor) Close() error {
	return p.subscriber.Close()
}

func main() {
	ctx := context.Background()

	nc, err := events.ConnectWithDefaults("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	subscriber, err := events.NewSubscriber(nc)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}

	processor := NewUserEventProcessor(subscriber)
	defer processor.Close()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to start event processor: %v", err)
	}

	log.Println("User event processor started successfully")
	log.Println("Listening for events on subject: auth.user.*")
	log.Println("Press Ctrl+C to stop...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}
