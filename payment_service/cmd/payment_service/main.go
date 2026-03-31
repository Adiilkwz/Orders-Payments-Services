package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"payment_service/internal/config"
	"payment_service/internal/repository"
	"payment_service/internal/transport/http"
	"payment_service/internal/usecase"
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
	log.Println("Successfully connected to the payments_db!")

	paymentRepo := repository.NewPostgresPaymentRepo(db)

	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	paymentHandler := http.NewPaymentHandler(paymentUC)

	r := gin.Default()

	r.POST("/payments", paymentHandler.ProcessPayment)
	r.GET("/payments/:order_id", paymentHandler.GetPaymentStatus)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting Payment Service on :%s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
