package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"order_service/internal/client"
	"order_service/internal/config"
	"order_service/internal/repository"
	"order_service/internal/transport/http"
	"order_service/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables")
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to the database!")

	paymentURL := os.Getenv("PAYMENT_SERVICE_URL")
	orderRepo := repository.NewPostgresOrderRepo(db)
	paymentGateway := client.NewPaymentHTTPClient(paymentURL)

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentGateway)

	orderHandler := http.NewOrderHandler(orderUC)

	r := gin.Default()
	r.POST("/orders", orderHandler.CreateOrder)
	r.GET("/orders/:id", orderHandler.GetOrder)
	r.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Order Service on :%s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
