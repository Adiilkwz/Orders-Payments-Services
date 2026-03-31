package usecase

import (
	"payment_service/internal/domain"

	"github.com/google/uuid"
)

type paymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *paymentUseCase {
	return &paymentUseCase{
		repo: repo,
	}
}

func (u *paymentUseCase) ProcessPayment(orderID string, amount int64) (*domain.Payment, error) {
	payment := &domain.Payment{
		ID:      uuid.New().String(),
		OrderID: orderID,
		Amount:  amount,
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

func (u *paymentUseCase) GetPaymentStatus(orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(orderID)
}
