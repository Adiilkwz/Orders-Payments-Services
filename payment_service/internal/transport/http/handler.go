package http

import (
	"net/http"

	"payment_service/internal/domain"

	"github.com/gin-gonic/gin"
)

type PaymentUseCase interface {
	ProcessPayment(orderID string, amount int64) (*domain.Payment, error)
	GetPaymentStatus(orderID string) (*domain.Payment, error)
}

type PaymentHandler struct {
	useCase PaymentUseCase
}

func NewPaymentHandler(uc PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{
		useCase: uc,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
		Amount  int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	payment, err := h.useCase.ProcessPayment(req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": payment.TransactionID,
		"status":         string(payment.Status),
	})
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.useCase.GetPaymentStatus(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found for this order"})
		return
	}

	c.JSON(http.StatusOK, payment)
}
