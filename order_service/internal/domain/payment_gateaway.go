package domain

type PaymentResult struct {
	TransactionID string
	Status        string
}

type PaymentGateway interface {
	ProcessPayment(orderID string, amount int64) (*PaymentResult, error)
}
