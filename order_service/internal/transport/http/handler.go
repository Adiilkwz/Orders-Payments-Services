package http

import (
	"log"
	"net/http"
	"strings"

	"order_service/internal/broker"
	"order_service/internal/domain"

	orderv1 "github.com/Adiilkwz/grpc-generated-go/order/v1"
	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	CustomerID    string `json:"customer_id" binding:"required"`
	CustomerEmail string `json:"customer_email" binding:"required"`
	ItemName      string `json:"item_name" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
}

type OrderUseCase interface {
	CreateOrder(customerID string, customerEmail string, itemName string, amount int64) (*domain.Order, error)
	GetByOrderID(id string) (*domain.Order, error)
	CancelOrder(id string) error
}

type OrderHandler struct {
	orderv1.UnimplementedOrderTrackingServiceServer
	useCase OrderUseCase
	hub     *broker.Hub
}

func NewOrderHandler(uc OrderUseCase, hub *broker.Hub) *OrderHandler {
	return &OrderHandler{
		useCase: uc,
		hub:     hub,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req struct {
		CustomerID    string `json:"customer_id" binding:"required"`
		CustomerEmail string `json:"customer_email" binding:"required"`
		ItemName      string `json:"item_name" binding:"required"`
		Amount        int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	order, err := h.useCase.CreateOrder(req.CustomerID, req.CustomerEmail, req.ItemName, req.Amount)
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

func (h *OrderHandler) SubscribeToOrderUpdates(req *orderv1.OrderRequest, stream orderv1.OrderTrackingService_SubscribeToOrderUpdatesServer) error {
	log.Printf("New subscriber to order status: %s", req.OrderId)

	events := h.hub.Subscribe(req.OrderId)

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("The customer canceled the order: %s", req.OrderId)
			return nil

		case event := <-events:
			err := stream.Send(&orderv1.OrderStatusUpdate{
				OrderId: event.OrderID,
				Status:  event.Status,
			})
			if err != nil {
				log.Printf("Error sending to the stream: %v", err)
				return err
			}
			log.Printf("The update has been sent to the stream for %s: %s", event.OrderID, event.Status)
		}
	}
}
