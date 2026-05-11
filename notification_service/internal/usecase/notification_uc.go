package usecase

import (
	"log"
	"sync"

	"notification_service/internal/domain"
)

type notificationUseCase struct {
	processedEvents sync.Map
}

func NewNotificationUseCase() domain.NotificationUseCase {
	return &notificationUseCase{}
}

func (u *notificationUseCase) ProcessPayment(event domain.PaymentCompletedEvent) error {
	_, alreadyProcessed := u.processedEvents.Load(event.OrderID)
	if alreadyProcessed {
		log.Printf("[Idempotency] Duplicate! Email for order #%s was already sent.", event.OrderID)
		return nil
	}

	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%d",
		event.CustomerEmail, event.OrderID, event.Amount)

	u.processedEvents.Store(event.OrderID, true)

	return nil
}
