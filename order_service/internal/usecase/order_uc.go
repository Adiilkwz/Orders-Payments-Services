package usecase

import (
	"errors"
	"log"
	"time"

	"order_service/internal/broker"
	"order_service/internal/domain"

	"github.com/google/uuid"
)

type orderUseCase struct {
	repo    domain.OrderRepository
	gateway domain.PaymentGateway
	hub     *broker.Hub
}

func NewOrderUseCase(repo domain.OrderRepository, gateway domain.PaymentGateway, hub *broker.Hub) *orderUseCase {
	return &orderUseCase{
		repo:    repo,
		gateway: gateway,
		hub:     hub,
	}
}

func (u *orderUseCase) CreateOrder(customerID string, customerEmail string, itemName string, amount int64) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("invalid order: amount must be greater than 0")
	}

	order := &domain.Order{
		ID:            uuid.New().String(),
		CustomerID:    customerID,
		ItemName:      itemName,
		Amount:        amount,
		CustomerEmail: customerEmail,
		Status:        domain.StatusPending,
		CreatedAt:     time.Now(),
	}

	err := u.repo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	log.Printf("➡️ 2. UseCase: Отправляем email в gRPC: '%s'", customerEmail)

	paymentResult, paymentErr := u.gateway.ProcessPayment(order.ID, order.Amount, order.CustomerEmail)

	finalStatus := domain.StatusFailed

	if paymentErr == nil && paymentResult != nil && paymentResult.Status == "Authorized" {
		finalStatus = domain.StatusPaid
	}

	updateErr := u.repo.UpdateStatus(order.ID, finalStatus)
	if updateErr != nil {
		return nil, updateErr
	}

	order.Status = finalStatus

	u.hub.Publish(broker.OrderEvent{
		OrderID: order.ID,
		Status:  string(order.Status),
	})

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

	err = u.repo.UpdateStatus(id, domain.StatusCancelled)
	if err != nil {
		return err
	}

	u.hub.Publish(broker.OrderEvent{
		OrderID: id,
		Status:  string(domain.StatusCancelled),
	})

	return nil
}
