package domain

type PaymentStatus string

const (
	StatusAuthorized PaymentStatus = "Authorized"
	StatusDeclined   PaymentStatus = "Declined"
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        PaymentStatus
}

type PaymentRepository interface {
	CreatePayment(payment *Payment) error
	GetByOrderID(order_id string) (*Payment, error)
}
