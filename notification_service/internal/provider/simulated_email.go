package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type SimulatedEmailSender struct{}

func NewSimulatedEmailSender() *SimulatedEmailSender {
	return &SimulatedEmailSender{}
}

func (s *SimulatedEmailSender) Send(to string, orderID string, amount int64) error {
	log.Printf("[Simulated API] Connecting to mail server for %s...", to)

	latency := time.Duration(rand.Intn(1500)+500) * time.Millisecond
	time.Sleep(latency)

	if rand.Float32() < 0.30 {
		return errors.New("503 Service Unavailable: simulated external provider error")
	}

	log.Printf("[Simulated API] Email successfully delivered to %s (Order: %s, Amount: %d)", to, orderID, amount)
	return nil
}
