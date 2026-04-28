package usecase

import (
	"time"

	"payment_service/internal/domain"

	"github.com/google/uuid"
)

type PaymentUseCase interface {
	ProcessPayment(orderID string, amount int64) (*domain.Payment, error)
	ListPayments(status string) ([]*domain.Payment, error)
}

type paymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) PaymentUseCase {
	return &paymentUseCase{
		repo: repo,
	}
}

func (u *paymentUseCase) ProcessPayment(orderID string, amount int64) (*domain.Payment, error) {
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

	return payment, nil
}

func (u *paymentUseCase) ListPayments(status string) ([]*domain.Payment, error) {
	return u.repo.ListByStatus(status)
}
