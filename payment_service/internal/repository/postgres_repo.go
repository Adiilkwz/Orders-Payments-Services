package repository

import (
	"database/sql"
	"errors"
	"time"

	"payment_service/internal/domain"
)

type postgresPaymentRepo struct {
	db *sql.DB
}

func NewPostgresPaymentRepo(db *sql.DB) domain.PaymentRepository {
	return &postgresPaymentRepo{
		db: db,
	}
}

func (r *postgresPaymentRepo) CreatePayment(payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	txID := sql.NullString{
		String: payment.TransactionID,
		Valid:  payment.TransactionID != "",
	}

	_, err := r.db.Exec(
		query,
		payment.ID,
		payment.OrderID,
		txID,
		payment.Amount,
		string(payment.Status),
		time.Now(),
	)

	return err
}

func (r *postgresPaymentRepo) GetByOrderID(orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments
		WHERE order_id = $1
	`

	row := r.db.QueryRow(query, orderID)

	var payment domain.Payment
	var statusStr string
	var txID sql.NullString

	err := row.Scan(
		&payment.ID,
		&payment.OrderID,
		&txID,
		&payment.Amount,
		&statusStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment record not found")
		}
		return nil, err
	}

	payment.Status = domain.PaymentStatus(statusStr)

	if txID.Valid {
		payment.TransactionID = txID.String
	} else {
		payment.TransactionID = ""
	}

	return &payment, nil
}
