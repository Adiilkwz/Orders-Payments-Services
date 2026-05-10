package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"order_service/internal/broker"
	"order_service/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type orderUseCase struct {
	repo    domain.OrderRepository
	gateway domain.PaymentGateway
	hub     *broker.Hub
	redis   *redis.Client
}

func NewOrderUseCase(repo domain.OrderRepository, gateway domain.PaymentGateway, hub *broker.Hub, redis *redis.Client) *orderUseCase {
	return &orderUseCase{
		repo:    repo,
		gateway: gateway,
		hub:     hub,
		redis:   redis,
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
	ctx := context.Background()
	cacheKey := fmt.Sprintf("order:%s", id)

	cachedData, err := u.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		var order domain.Order
		if err := json.Unmarshal([]byte(cachedData), &order); err == nil {
			log.Printf("[Cache HIT] Order %s was taken from Redis", id)
			return &order, nil
		}
		log.Printf("Cache parsing error %s: %v", id, err)
	} else if err != redis.Nil {
		log.Printf("Redis connection error: %v", err)
	}

	order, err := u.repo.GetOrderById(id)
	if err != nil {
		return nil, err
	}

	orderJSON, marshalErr := json.Marshal(order)
	if marshalErr == nil {
		ttl := 5 * time.Minute
		if setErr := u.redis.Set(ctx, cacheKey, orderJSON, ttl).Err(); setErr != nil {
			log.Printf("Failed to save order %s into cache: %v", id, setErr)
		} else {
			log.Printf("[Cache MISS] Order %s was taken from DB and saved to Redis", id)
		}
	}

	return order, nil
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

	ctx := context.Background()
	cacheKey := fmt.Sprintf("order:%s", id)

	if err := u.redis.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("Failed to delete cache for order %s: %v", id, err)
	} else {
		log.Printf("[Cache INVALIDATED] Cache for order %s cleared due to cancellation", id)
	}

	u.hub.Publish(broker.OrderEvent{
		OrderID: id,
		Status:  string(domain.StatusCancelled),
	})

	return nil
}
