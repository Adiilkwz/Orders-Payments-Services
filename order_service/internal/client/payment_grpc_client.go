package client

import (
	"context"
	"fmt"
	"time"

	"order_service/internal/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	paymentv1 "github.com/Adiilkwz/grpc-generated-go/payment/v1"
)

type PaymentGRPCClient struct {
	client paymentv1.PaymentServiceClient
}

func NewPaymentGRPCClient(targetAddress string) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(targetAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %v", err)
	}

	c := paymentv1.NewPaymentServiceClient(conn)

	return &PaymentGRPCClient{client: c}, nil
}

func (p *PaymentGRPCClient) ProcessPayment(orderID string, amount int64) (*domain.PaymentResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &paymentv1.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	}

	res, err := p.client.ProcessPayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("grpc call failed: %v", err)
	}

	return &domain.PaymentResult{
		TransactionID: res.TransactionId,
		Status:        res.Status,
	}, nil
}
