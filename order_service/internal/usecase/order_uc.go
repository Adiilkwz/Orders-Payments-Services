package usecase

import (
	"errors"
	"time"

	"order_service/internal/domain"

	"github.com/google/uuid"
)

type orderUseCase struct {
	repo    domain.OrderRepository
	gateway domain.PaymentGateway
}

func NewOrderUseCase(repo domain.OrderRepository, gateway domain.PaymentGateway) *orderUseCase {
	return &orderUseCase{
		repo:    repo,
		gateway: gateway,
	}
}

func (u *orderUseCase) CreateOrder(customerID string, itemName string, amount int64) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("invalid order: amount must be greater than 0")
	}

	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     domain.StatusPending,
		CreatedAt:  time.Now(),
	}

	err := u.repo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	paymentResult, paymentErr := u.gateway.ProcessPayment(order.ID, order.Amount)

	finalStatus := domain.StatusFailed

	if paymentErr == nil && paymentResult != nil && paymentResult.Status == "Authorized" {
		finalStatus = domain.StatusPaid
	}

	updateErr := u.repo.UpdateStatus(order.ID, finalStatus)
	if updateErr != nil {
		return nil, updateErr
	}

	order.Status = finalStatus

	return order, nil
}

func (u *orderUseCase) GetByOrderID(id string) (*domain.Order, error) {
	return u.repo.GetOrderById(id)
}

func (u *orderUseCase) CancelOrder(id string) error {
	order, err := u.repo.GetOrderById(id)
	if err != nil {
		return err
	}

	if order.Status != domain.StatusPending {
		return errors.New("business rule violation: only 'pending' orders can be cancelled.")
	}

	return u.repo.UpdateStatus(id, domain.StatusCancelled)
}
