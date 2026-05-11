package domain

type PaymentCompletedEvent struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type EmailProvider interface {
	Send(to string, orderID string, amount int64) error
}

type NotificationUseCase interface {
	ProcessPayment(event PaymentCompletedEvent) error
}
