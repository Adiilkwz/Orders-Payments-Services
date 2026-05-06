# Project Overview & Message Broker Strategy

This project implements a microservices architecture using Go, PostgreSQL, gRPC, and RabbitMQ. A core requirement of distributed event-driven systems is ensuring that messages are reliably delivered and processed exactly once, even in the event of network failures or service restarts.

Below is the explanation of the strategies implemented in the Notification Service to handle these challenges.

## 1. Idempotency Strategy

Message brokers like RabbitMQ guarantee **"at-least-once"** delivery. This means that under certain network conditions, the same message might be delivered to the consumer more than once. If a payment notification is processed twice, the customer might receive duplicate emails, which is bad for user experience.

To solve this, we implemented an **Idempotency Check** in the Notification Service using Go's concurrent-safe `sync.Map`:

* **In-Memory Cache**: We store the `OrderID` of every successfully processed message in a `sync.Map` (`processedEvents`).
* **Validation**: Before processing a new message from the queue, the service checks: `_, alreadyProcessed := processedEvents.Load(event.OrderID)`.
* **Handling Duplicates**: If the `OrderID` is already in the map, the service recognizes it as a duplicate, logs a warning (`[Idempotency] Duplicate!`), skips processing, and safely discards the message.

## 2. Manual Acknowledgment (ACK) Logic

By default, RabbitMQ uses "auto-acknowledgment," meaning it deletes a message from the queue the moment it sends it to a consumer. If the consumer crashes before finishing its work, the message is lost forever.

To prevent data loss, we implemented **Manual Acknowledgments**:

* **Auto-Ack Disabled**: When declaring the consumer (`ch.Consume`), the `autoAck` parameter is set to `false`. This forces RabbitMQ to keep the message in the queue until the service explicitly says it is done.
* **Success ACK**: Only *after* the email logic is executed and the idempotency map is updated, the service calls `d.Ack(false)`. This tells RabbitMQ it is safe to delete the message.
* **Error ACK**: If the JSON unmarshalling fails or a message is identified as a duplicate, we still call `d.Ack(false)`. This prevents "poison messages" (messages that always cause errors) from being requeued infinitely and blocking the system.