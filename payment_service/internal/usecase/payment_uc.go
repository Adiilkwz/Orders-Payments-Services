package usecase

import (
	"log"
	"time"

	"payment_service/internal/broker"
	"payment_service/internal/domain"

	"github.com/google/uuid"
)

type PaymentUseCase interface {
	ProcessPayment(orderID string, amount int64, customerEmail string) (*domain.Payment, error)
	ListPayments(status string) ([]*domain.Payment, error)
}

type paymentUseCase struct {
	repo      domain.PaymentRepository
	publisher *broker.RabbitMQPublisher
}

func NewPaymentUseCase(repo domain.PaymentRepository, pub *broker.RabbitMQPublisher) PaymentUseCase {
	return &paymentUseCase{
		repo:      repo,
		publisher: pub,
	}
}

func (u *paymentUseCase) ProcessPayment(orderID string, amount int64, customerEmail string) (*domain.Payment, error) {
	payment := &domain.Payment{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Amount:    amount,
		CreatedAt: time.Now(),
	}

	if amount > 100000 {
		payment.Status = domain.StatusDeclined
	} else {
		payment.Status = domain.StatusAuthorized
		payment.TransactionID = uuid.New().String()
	}

	err := u.repo.CreatePayment(payment)
	if err != nil {
		return nil, err
	}

	log.Printf("➡️ 4. Publisher: Кладем email в RabbitMQ: '%s'", customerEmail)

	if payment.Status == domain.StatusAuthorized {
		event := broker.PaymentCompletedEvent{
			OrderID:       orderID,
			Amount:        amount,
			CustomerEmail: customerEmail,
			Status:        string(payment.Status),
		}

		go func() {
			_ = u.publisher.PublishPaymentCompleted(event)
		}()
	}

	return payment, nil
}

func (u *paymentUseCase) ListPayments(status string) ([]*domain.Payment, error) {
	return u.repo.ListByStatus(status)
}
