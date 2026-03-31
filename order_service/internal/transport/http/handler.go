package http

import (
	"net/http"
	"strings"

	"order_service/internal/domain"

	"github.com/gin-gonic/gin"
)

type OrderUseCase interface {
	CreateOrder(customerID string, itemName string, amount int64) (*domain.Order, error)
	GetByOrderID(id string) (*domain.Order, error)
	CancelOrder(id string) error
}

type OrderHandler struct {
	useCase OrderUseCase
}

func NewOrderHandler(uc OrderUseCase) *OrderHandler {
	return &OrderHandler{
		useCase: uc,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id" binding:"required"`
		ItemName   string `json:"item_name" binding:"required"`
		Amount     int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	order, err := h.useCase.CreateOrder(req.CustomerID, req.ItemName, req.Amount)
	if err != nil {
		if strings.Contains(err.Error(), "payment service unavailable") || strings.Contains(err.Error(), "timed out") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Payment Service is currently unavailable. Order failed."})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.useCase.GetByOrderID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	err := h.useCase.CancelOrder(id)
	if err != nil {
		if strings.Contains(err.Error(), "business rule violation") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}
