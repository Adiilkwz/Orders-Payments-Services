# Microservice Order & Payment System

This project implements a distributed e-commerce backend using **Golang** and **Clean Architecture**. It consists of two independent microservices communicating via HTTP REST APIs, each managing its own strictly isolated PostgreSQL database.

## Architecture Decisions

### 1. Clean Architecture
Each service is heavily decoupled into four distinct layers to ensure the business logic remains entirely independent of external frameworks (like Gin) or databases (PostgreSQL):
* **Domain:** Defines the core entities (`Order`, `Payment`) and repository interfaces.
* **Use Case:** Contains the pure business logic and enforces state invariants.
* **Repository:** Implements data persistence (PostgreSQL adapter).
* **Transport:** Handles HTTP requests and responses (Gin adapter).

### 2. Bounded Contexts & Database Isolation
The system is divided into two Bounded Contexts to ensure high cohesion and loose coupling:
* **Order Context:** Manages the customer's purchase lifecycle.
* **Payment Context:** Solely responsible for authorizing financial transactions and enforcing payment limits.
* **Database-per-Service:** To enforce true microservice isolation, the Order Service connects only to `orders_db`, and the Payment Service connects only to `payments_db`. They do not share tables or database credentials.

---

## Business Rules & Invariants

The Use Case layer strictly enforces the following domain rules:
1. **Financial Accuracy:** All monetary values are represented as `int64` (cents/minor units) to prevent floating-point precision errors.
2. **Order Validation:** Order amounts must be strictly `> 0`.
3. **State Machine Limits:** An order can only be Cancelled if it is in the "Pending" state. Once an order is "Paid", cancellation is strictly prohibited by the domain logic.
4. **Payment Limits:** The Payment Service enforces a hard limit: any transaction `> 100,000` units is automatically marked as "Declined".

---

## Failure Handling & Resilience

A core requirement of this distributed system is surviving network failures gracefully.

* **Timeouts:** The Order Service communicates with the Payment Service using a custom `http.Client` configured with a strict **2-second timeout**.
* **Circuit Breaking:** If the Payment Service is offline, unreachable, or times out, the Order Service does not hang indefinitely. It immediately catches the error and returns a clean `503 Service Unavailable` to the client.
* **Compensating Action:** In the event of a Payment Service failure, the Order state is explicitly updated to **"Failed"** in the database. 
  * *Defense Justification:* Marking it as "Failed" rather than leaving it "Pending" provides immediate clarity to the user and prevents edge cases where a background job might accidentally retry and double-charge a stagnant pending order later.

---

## How to Run (Local Environment)

This project is configured to run the Go services locally while connecting to a manual PostgreSQL instance (e.g., via pgAdmin).

### 1. Database Setup
1. Create two databases in PostgreSQL: `orders_db` and `payments_db`.
2. Execute the migration scripts located in `order_service/migrations/` and `payment_service/migrations/` to create the tables and ENUM types.

### 2. Start the Services
Open two separate terminals:

**Terminal 1 (Payment Service - Port 8081):**
```bash
cd payment_service
go run cmd/payment_service/main.go