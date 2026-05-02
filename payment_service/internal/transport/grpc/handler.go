package grpc

import (
	"context"
	"time"

	"payment_service/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentv1 "github.com/Adiilkwz/grpc-generated-go/payment/v1"
)

type PaymentHandler struct {
	paymentv1.UnimplementedPaymentServiceServer
	useCase PaymentUseCaseInterface
}

type PaymentUseCaseInterface interface {
	ProcessPayment(orderID string, amount int64, customerEmail string) (*domain.Payment, error)
	ListPayments(status string) ([]*domain.Payment, error)
}

func NewPaymentHandler(uc PaymentUseCaseInterface) *PaymentHandler {
	return &PaymentHandler{useCase: uc}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be greater than 0")
	}

	payment, err := h.useCase.ProcessPayment(req.OrderId, req.Amount, req.CustomerEmail)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment processing failed: %v", err)
	}

	return &paymentv1.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        string(payment.Status),
		ProcessedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *paymentv1.ListPaymentRequest) (*paymentv1.ListPaymentResponse, error) {
	payments, err := h.useCase.ListPayments(req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch payments: %v", err)
	}

	var pbPayments []*paymentv1.PaymentResponse
	for _, p := range payments {
		pbPayments = append(pbPayments, &paymentv1.PaymentResponse{
			TransactionId: p.ID,
			Status:        string(p.Status),
			ProcessedAt:   timestamppb.New(p.CreatedAt),
		})
	}

	return &paymentv1.ListPaymentResponse{
		Payments: pbPayments,
	}, nil
}
