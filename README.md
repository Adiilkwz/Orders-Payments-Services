# Project Overview & Message Broker Strategy

This project implements a microservices architecture using Go, PostgreSQL, Redis, gRPC, and RabbitMQ. A core requirement of distributed event-driven systems is ensuring that messages are reliably delivered, processed exactly once, and robust against external failures.

Below is the explanation of the strategies implemented to handle these challenges at scale.

## 1. Distributed Idempotency Strategy

Message brokers like RabbitMQ guarantee **"at-least-once"** delivery. This means that under certain network conditions, the same message might be delivered to the consumer more than once. 

To prevent duplicate emails, we implemented an **Idempotency Check** in the Notification Service using **Redis**:
* **Distributed Lock**: Before processing a message, the service uses the Redis `SETNX` (Set if Not eXists) command with a unique key (`notification:payment:{order_id}`).
* **Validation**: If `SETNX` returns `true`, it's a new message, and processing continues. If it returns `false`, the message is identified as a duplicate, logged as a blocked duplicate, and safely discarded.
* **Why Redis?**: Unlike an in-memory map (`sync.Map`), Redis provides a centralized, persistent state. If the Notification Service restarts or scales horizontally to multiple instances, the idempotency check remains fully intact.

## 2. Manual Acknowledgment (ACK) Logic

By default, RabbitMQ uses "auto-acknowledgment," meaning it deletes a message from the queue the moment it sends it to a consumer. If the consumer crashes before finishing its work, the message is lost forever.

To prevent data loss, we implemented **Manual Acknowledgments**:
* **Auto-Ack Disabled**: When declaring the consumer, `autoAck` is set to `false`. RabbitMQ keeps the message until explicitly told otherwise.
* **Safe ACK**: The `d.Ack(false)` function is only called *after* the email logic has fully succeeded, or if the message has permanently failed all retries. This ensures no message is lost due to an unexpected panic or crash mid-process.

## 3. Caching & Invalidation Strategy (Cache-Aside)

To reduce the load on the primary PostgreSQL database, the Order Service utilizes Redis implementing the **Cache-Aside Pattern**:
* **Read Path (GET)**: When fetching an order, the system first queries Redis. If found (**Cache Hit**), it returns instantly. If missing (**Cache Miss**), it queries PostgreSQL, returns the data to the user, and asynchronously saves the result in Redis with a 5-minute TTL (Time-To-Live).
* **Write Path / Invalidation**: Caching introduces the risk of serving stale data. To solve this, whenever an order's state changes (e.g., an order is canceled), the service executes an explicit Cache Invalidation (`redis.Del`) to delete the stale key. The next request will be forced to fetch the fresh data from the database.

## 4. Retry Logic & Exponential Backoff

External providers (like email APIs) can experience transient network issues or downtime. If the Notification Service fails to send an email, it should not give up immediately, nor should it spam the provider.
* **The Retry Loop**: The Adapter pattern wraps the external provider in a loop allowing up to 3 retry attempts.
* **Exponential Backoff**: If a `503 Service Unavailable` error occurs, the worker pauses execution using an exponentially increasing delay ($2^n$ seconds: 2s, 4s, 8s) before trying again. 
* **Poison Messages**: If the request fails after the maximum retries, the job is marked as "failed" in Redis and the message is manually ACKed to prevent it from indefinitely blocking the queue.