package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"payment_service/internal/broker"
	"payment_service/internal/config"
	"payment_service/internal/repository"
	transport "payment_service/internal/transport/grpc"
	"payment_service/internal/usecase"

	paymentv1 "github.com/Adiilkwz/grpc-generated-go/payment/v1"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it")
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	rabbitPublisher, err := broker.NewRabbitMQPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitPublisher.Close()

	repo := repository.NewPostgresPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo, rabbitPublisher)

	grpcHandler := transport.NewPaymentHandler(uc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcHandler)

	log.Printf("Payment gRPC Service is securely running on port %s...", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
