package main

import (
	"context"
	"log"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/events"
)

type UserService struct {
	publisher   events.Publisher
	timeManager TimeManager
}

type TimeManager interface {
	Now() time.Time
}

type RealTimeManager struct{}

func (t *RealTimeManager) Now() time.Time {
	return time.Now().UTC()
}

func NewUserService(publisher events.Publisher, timeManager TimeManager) *UserService {
	return &UserService{
		publisher:   publisher,
		timeManager: timeManager,
	}
}

func (s *UserService) CreateUser(ctx context.Context, userID int64, email string, role string) error {
	event := events.NewEvent(
		"user.created",
		"auth-service",
		"org-123",
		s.timeManager.Now(),
		map[string]interface{}{
			"user_id": userID,
			"email":   email,
			"role":    role,
		},
	)

	if err := s.publisher.Publish(ctx, "auth.user.created", event); err != nil {
		log.Printf("Failed to publish user.created event: %v", err)
		return err
	}

	log.Printf("Published user.created event for user %d", userID)
	return nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, userID int64, oldRole, newRole string) error {
	event := events.NewEvent(
		"user.role.updated",
		"auth-service",
		"org-123",
		s.timeManager.Now(),
		map[string]interface{}{
			"user_id":  userID,
			"old_role": oldRole,
			"new_role": newRole,
		},
	)

	if err := s.publisher.PublishAsync(ctx, "auth.user.role.updated", event); err != nil {
		log.Printf("Failed to publish user.role.updated event: %v", err)
		return err
	}

	log.Printf("Published user.role.updated event for user %d", userID)
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID int64, reason string) error {
	event := events.NewEvent(
		"user.deleted",
		"auth-service",
		"org-123",
		s.timeManager.Now(),
		map[string]interface{}{
			"user_id": userID,
			"reason":  reason,
		},
	)

	if err := s.publisher.Publish(ctx, "auth.user.deleted", event); err != nil {
		log.Printf("Failed to publish user.deleted event: %v", err)
		return err
	}

	log.Printf("Published user.deleted event for user %d", userID)
	return nil
}

func (s *UserService) Close() error {
	return s.publisher.Close()
}

func main() {
	ctx := context.Background()

	nc, err := events.ConnectWithDefaults("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	publisher, err := events.NewPublisher(nc)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}

	timeManager := &RealTimeManager{}
	userService := NewUserService(publisher, timeManager)
	defer userService.Close()

	if err := userService.CreateUser(ctx, 12345, "user@example.com", "admin"); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	if err := userService.UpdateUserRole(ctx, 12345, "admin", "user"); err != nil {
		log.Fatalf("Failed to update user role: %v", err)
	}

	if err := userService.DeleteUser(ctx, 12345, "user_requested"); err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}

	log.Println("All events published successfully")
}
