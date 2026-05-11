# Microservices Architecture Diagram

This diagram illustrates the full lifecycle of the system, including HTTP requests, Cache-Aside logic, synchronous gRPC communication, asynchronous message queues, and robust background worker processing with exponential backoff.

```mermaid
sequenceDiagram
    participant Client as Postman / Client
    participant Order as Order Service
    participant OrderDB as PostgreSQL (orders_db)
    participant Redis as Redis (Cache & Idempotency)
    participant Payment as Payment Service
    participant PaymentDB as PostgreSQL (payments_db)
    participant RabbitMQ as RabbitMQ Broker
    participant Notification as Notification Worker
    participant ExternalAPI as External Provider (Adapter)
    
    %% Cache-Aside Pattern Flow (Read)
    Note over Client, Redis: Phase 2: Cache-Aside Pattern (GET /orders/:id)
    Client->>Order: GET /orders/:id (HTTP)
    activate Order
    Order->>Redis: Get "order:id"
    alt Cache Hit
        Redis-->>Order: Return cached JSON
    else Cache Miss
        Redis-->>Order: nil
        Order->>OrderDB: Fetch from DB
        OrderDB-->>Order: Success
        Order->>Redis: Set "order:id" (TTL: 5m)
    end
    Order-->>Client: 200 OK
    deactivate Order

    %% Main Creation Flow
    Note over Client, Notification: Phase 1 & 3: Order Creation & Event Publishing
    Client->>Order: POST /orders (HTTP)
    activate Order
    Order->>OrderDB: Save Order Details
    Order->>Payment: gRPC: ProcessPayment(OrderID, Amount)
    activate Payment
    Payment->>PaymentDB: Save Payment Status
    Payment->>RabbitMQ: Publish Event (payment.completed)
    Payment-->>Order: gRPC Response (Success)
    deactivate Payment
    Order-->>Client: 200 OK (Order Created)
    deactivate Order
    
    %% Background Job Flow with Retry
    Note over RabbitMQ, ExternalAPI: Phase 4: Reliable Background Job
    RabbitMQ-->>Notification: Push Event (payment.completed)
    activate Notification
    
    Notification->>Redis: SETNX idempotency_key (Check Duplicate)
    Redis-->>Notification: OK (Is New Message)
    
    loop Exponential Backoff Retry (Max 3)
        Notification->>ExternalAPI: Send Email
        alt Simulated Network Failure
            ExternalAPI-->>Notification: 503 Service Unavailable
            Notification->>Notification: time.Sleep (2s, 4s, 8s)
        else Success
            ExternalAPI-->>Notification: 200 OK
        end
    end
    
    Notification->>Redis: Update key status to "completed"
    Notification->>RabbitMQ: Acknowledge Message (Manual ACK)
    deactivate Notification