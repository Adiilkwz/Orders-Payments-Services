package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

var processedEvents sync.Map

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	amqpURL := os.Getenv("AMQP_URL")

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to initialize queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("[Notification Service] started and are waiting for message. Press Ctrl+C for exit.")

	go func() {
		for d := range msgs {
			var event PaymentCompletedEvent

			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Printf("JSON parsing error: %v", err)
				d.Ack(false)
				continue
			}

			_, alreadyProcessed := processedEvents.Load(event.OrderID)

			if alreadyProcessed {
				log.Printf("[Idempotency] Duplicate! Email for order #%s was already sent.", event.OrderID)
				d.Ack(false)
				continue
			}

			log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%d",
				event.CustomerEmail, event.OrderID, event.Amount)

			processedEvents.Store(event.OrderID, true)

			d.Ack(false)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	log.Println("Completion of Notification Service...")
}
