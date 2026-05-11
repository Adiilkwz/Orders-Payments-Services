package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification_service/internal/broker"
	"notification_service/internal/domain"
	"notification_service/internal/provider"
	"notification_service/internal/usecase"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	amqpURL := os.Getenv("AMQP_URL")

	var emailProvider domain.EmailProvider
	providerMode := os.Getenv("PROVIDER_MODE")

	if providerMode == "REAL" {
		log.Println("Initialized REAL email provider")
	} else {
		emailProvider = provider.NewSimulatedEmailSender()
		log.Println("Initialized SIMULATED email provider (Expect random failures!)")
	}

	uc := usecase.NewNotificationUseCase(emailProvider)

	consumer, err := broker.NewConsumer(amqpURL, uc)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("Graceful shutdown of Notification Service...")
}
