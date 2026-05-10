package main

import (
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"order_service/internal/broker"
	"order_service/internal/client"
	"order_service/internal/config"
	"order_service/internal/repository"
	"order_service/internal/transport/http"
	"order_service/internal/usecase"

	orderv1 "github.com/Adiilkwz/grpc-generated-go/order/v1"
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

	redis, err := config.ConnectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	paymentTarget := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentTarget == "" {
		paymentTarget = "localhost:50051"
	}

	paymentGRPCClient, err := client.NewPaymentGRPCClient(paymentTarget)
	if err != nil {
		log.Fatalf("Failed to start gRPC client: %v", err)
	}
	log.Printf("Connected to Payment Service via gRPC at %s", paymentTarget)

	hub := broker.NewHub()

	repo := repository.NewPostgresOrderRepo(db)

	uc := usecase.NewOrderUseCase(repo, paymentGRPCClient, hub, redis)
	handler := http.NewOrderHandler(uc, hub)

	grpcPort := "50052"
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderTrackingServiceServer(grpcServer, handler)

	go func() {
		log.Printf("Order Tracking gRPC Stream Server running on port %s", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	r := gin.Default()

	r.POST("/orders", handler.CreateOrder)
	r.POST("/orders/:id/cancel", handler.CancelOrder)
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
