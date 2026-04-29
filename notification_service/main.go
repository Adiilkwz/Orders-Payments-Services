package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	log.Println("⏳ [Notification Service] started and are waiting for message. Press Ctrl+C for exit.")

	go func() {
		for d := range msgs {
			var event PaymentCompletedEvent

			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Printf("JSON parsing error: %v", err)
				d.Ack(false)
				continue
			}

			log.Printf("📧 [Notification] Sent email to %s for Order #%s. Amount: $%d",
				event.CustomerEmail, event.OrderID, event.Amount)

			d.Ack(false)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	log.Println("Completion of Notification Service...")
}
