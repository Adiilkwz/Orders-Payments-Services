package domain

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "Pending"
	StatusPaid      OrderStatus = "Paid"
	StatusFailed    OrderStatus = "Failed"
	StatusCancelled OrderStatus = "Cancelled"
)

type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64
	Status     OrderStatus
	CreatedAt  time.Time
}

type OrderRepository interface {
	Create(*Order) error
	GetOrderById(id string) (*Order, error)
	UpdateStatus(id string, status OrderStatus) error
}
