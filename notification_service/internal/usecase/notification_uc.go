package usecase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"notification_service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type notificationUseCase struct {
	processedEvents sync.Map
	emailProvider   domain.EmailProvider
	redis           *redis.Client
}

func NewNotificationUseCase(emailProvider domain.EmailProvider, redisClient *redis.Client) domain.NotificationUseCase {
	return &notificationUseCase{
		emailProvider: emailProvider,
		redis:         redisClient,
	}
}

func (u *notificationUseCase) ProcessPayment(event domain.PaymentCompletedEvent) error {
	ctx := context.Background()
	idempotencyKey := "notification:payment:" + event.OrderID

	isNew, err := u.redis.SetNX(ctx, idempotencyKey, "processing", 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("redis connection error: %v", err)
	}
	if !isNew {
		log.Printf("🛡️ [Idempotency] Duplicate blocked by Redis! Email for order #%s was already processed.", event.OrderID)
		return nil
	}

	maxRetries := 3
	var sendErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		sendErr = u.emailProvider.Send(event.CustomerEmail, event.OrderID, event.Amount)

		if sendErr == nil {
			u.redis.Set(ctx, idempotencyKey, "completed", 24*time.Hour)
			return nil
		}

		log.Printf("Send failed (Attempt %d/%d): %v", attempt, maxRetries, sendErr)

		if attempt < maxRetries {
			backoff := time.Duration(1<<attempt) * time.Second
			log.Printf("Network unstable. Waiting %v before retrying...", backoff)
			time.Sleep(backoff)
		}
	}

	u.redis.Set(ctx, idempotencyKey, "failed", 24*time.Hour)
	log.Printf("Permanently failed to send email for order #%s after %d attempts", event.OrderID, maxRetries)

	return nil
}
