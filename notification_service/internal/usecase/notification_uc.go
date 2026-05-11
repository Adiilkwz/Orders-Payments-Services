package usecase

import (
	"log"
	"sync"

	"notification_service/internal/domain"
)

type notificationUseCase struct {
	processedEvents sync.Map
	emailProvider   domain.EmailProvider
}

func NewNotificationUseCase(emailProvider domain.EmailProvider) domain.NotificationUseCase {
	return &notificationUseCase{
		emailProvider: emailProvider,
	}
}

func (u *notificationUseCase) ProcessPayment(event domain.PaymentCompletedEvent) error {
	_, alreadyProcessed := u.processedEvents.Load(event.OrderID)
	if alreadyProcessed {
		log.Printf("[Idempotency] Duplicate! Email for order #%s was already sent.", event.OrderID)
		return nil
	}

	err := u.emailProvider.Send(event.CustomerEmail, event.OrderID, event.Amount)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%d",
		event.CustomerEmail, event.OrderID, event.Amount)

	u.processedEvents.Store(event.OrderID, true)

	return nil
}
