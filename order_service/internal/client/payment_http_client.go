package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"order_service/internal/domain"
)

type paymentHTTPClient struct {
	client  *http.Client
	baseURL string
}

func NewPaymentHTTPClient(baseURL string) domain.PaymentGateway {
	return &paymentHTTPClient{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *paymentHTTPClient) ProcessPayment(orderID string, amount int64) (*domain.PaymentResult, error) {
	payload := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("failed to marshal payment request")
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.New("payment service unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, errors.New("payment service unavailable")
	}

	var result struct {
		TransactionID string `json:"transaction_id"`
		Status        string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.New("failed to decode payment response")
	}

	return &domain.PaymentResult{
		TransactionID: result.TransactionID,
		Status:        result.Status,
	}, nil
}
