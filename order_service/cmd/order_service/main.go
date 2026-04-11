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
		log.Println("Warning: No .env file found")
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	paymentTarget := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentTarget == "" {
		paymentTarget = "localhost:50051"
	}

	paymentGRPCClient, err := client.NewPaymentGRPCClient(paymentTarget)
	if err != nil {
		log.Fatalf("Failed to start gRPC client: %v", err)
	}
	log.Printf("Connected to Payment Service via gRPC at %s", paymentTarget)

	repo := repository.NewPostgresOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo, paymentGRPCClient)
	handler := http.NewOrderHandler(uc)

	r := gin.Default()

	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order Service REST API running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}
