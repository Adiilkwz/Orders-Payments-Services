package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	orderv1 "github.com/Adiilkwz/grpc-generated-go/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := orderv1.NewOrderTrackingServiceClient(conn)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <ORDER_ID>")
		return
	}
	orderID := os.Args[1]

	fmt.Printf("Connecting to stream for ordering: %s...\n", orderID)

	stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderv1.OrderRequest{OrderId: orderID})
	if err != nil {
		log.Fatalf("Subscribe error: %v", err)
	}

	fmt.Println("Stream is opened! Waiting on status update...")
	fmt.Println("-------------------------------------------")

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream is locked by the server")
			break
		}
		if err != nil {
			log.Fatalf("Failed to read from stream: %v", err)
		}

		fmt.Printf("New Status: Order %s now on status [%s]\n", msg.OrderId, msg.Status)
	}
}
