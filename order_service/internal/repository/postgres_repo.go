package repository

import (
	"database/sql"
	"errors"

	"order_service/internal/domain"
)

type postgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) domain.OrderRepository {
	return &postgresOrderRepo{
		db: db,
	}
}

func (r *postgresOrderRepo) CreateOrder(order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		string(order.Status),
		order.CreatedAt,
	)

	return err
}

func (r *postgresOrderRepo) GetOrderById(id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var order domain.Order
	var statusStr string

	err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&statusStr,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	order.Status = domain.OrderStatus(statusStr)

	return &order, nil
}

func (r *postgresOrderRepo) UpdateStatus(id string, status domain.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, string(status), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("order not found or status unchanged")
	}

	return nil
}
