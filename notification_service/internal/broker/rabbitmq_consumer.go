package broker

import (
	"encoding/json"
	"log"
	"time"

	"notification_service/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	uc   domain.NotificationUseCase
}

func NewConsumer(amqpURL string, uc domain.NotificationUseCase) (*Consumer, error) {
	var conn *amqp.Connection
	var err error

	for i := 1; i <= 20; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ!")
			break
		}
		log.Printf("RabbitMQ is not ready (attempt %d/20). Waiting 5 seconds...", i)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{conn: conn, ch: ch, uc: uc}, nil
}

func (c *Consumer) Start() error {
	q, err := c.ch.QueueDeclare("payment.completed", true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := c.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("[Notification Service] started and waiting for messages. Press Ctrl+C to exit.")

	go func() {
		for d := range msgs {
			var event domain.PaymentCompletedEvent

			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("JSON parsing error: %v", err)
				d.Ack(false)
				continue
			}

			err := c.uc.ProcessPayment(event)
			if err != nil {
				log.Printf("Error processing payment: %v", err)
			}

			d.Ack(false)
		}
	}()

	return nil
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
