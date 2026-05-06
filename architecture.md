# Microservices Architecture Diagram

This diagram illustrates the flow of a single order request through the system, demonstrating the synchronous gRPC communication and asynchronous event-driven communication via RabbitMQ.
```mermaid
sequenceDiagram
    participant Client as Postman / Client
    participant Order as Order Service (Port 8080)
    participant OrderDB as PostgreSQL (orders_db)
    participant Payment as Payment Service (Port 50051)
    participant PaymentDB as PostgreSQL (payments_db)
    participant RabbitMQ as RabbitMQ Broker (Port 5672)
    participant Notification as Notification Service
    
    %% Step 1: HTTP Request
    Client->>Order: POST /orders (HTTP)
    activate Order
    
    %% Step 2: Database Save
    Order->>OrderDB: Save Order Details
    OrderDB-->>Order: Success
    
    %% Step 3: Synchronous gRPC Call
    Order->>Payment: gRPC: ProcessPayment(OrderID, Amount)
    activate Payment
    
    %% Step 4: Database Save
    Payment->>PaymentDB: Save Payment Status
    PaymentDB-->>Payment: Success
    
    %% Step 5: Asynchronous Event Publishing
    Payment->>RabbitMQ: Publish Event (payment.completed)
    
    %% Step 6: gRPC Response
    Payment-->>Order: gRPC Response (Success)
    deactivate Payment
    
    %% Step 7: HTTP Response
    Order-->>Client: 200 OK (Order Created)
    deactivate Order
    
    %% Step 8: Asynchronous Consumption
    RabbitMQ-->>Notification: Push Event (payment.completed)
    activate Notification
    Notification->>Notification: Check Idempotency (sync.Map)
    Notification->>Notification: Acknowledge Message (Manual ACK)
    Notification->>Notification: Send Email/Log Result
    deactivate Notification