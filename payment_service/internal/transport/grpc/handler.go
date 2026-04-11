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
	ProcessPayment(orderID string, amount int64) (*domain.Payment, error)
	GetPaymentStatus(orderID string) (*domain.Payment, error)
}

func NewPaymentHandler(uc PaymentUseCaseInterface) *PaymentHandler {
	return &PaymentHandler{useCase: uc}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be greater than 0")
	}

	payment, err := h.useCase.ProcessPayment(req.OrderId, req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment processing failed: %v", err)
	}

	return &paymentv1.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        string(payment.Status),
		ProcessedAt:   timestamppb.New(time.Now()),
	}, nil
}
