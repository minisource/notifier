package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/minisource/go-sdk/notifier/proto/notifier/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to notifier gRPC
	conn, err := grpc.NewClient("localhost:9003", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewNotificationServiceClient(conn)

	// Test sending SMS with tokens
	req := &pb.SendSMSRequest{
		To:       "+989011793041",
		Template: "verify",
		Tokens: map[string]string{
			"code": "999888",
		},
	}

	fmt.Printf("Sending request:\n")
	fmt.Printf("  To: %s\n", req.To)
	fmt.Printf("  Template: %s\n", req.Template)
	fmt.Printf("  Tokens: %v\n", req.Tokens)
	fmt.Printf("  GetTokens(): %v\n", req.GetTokens())

	resp, err := client.SendSMS(context.Background(), req)
	if err != nil {
		log.Fatalf("SendSMS failed: %v", err)
	}

	fmt.Printf("\nResponse:\n")
	fmt.Printf("  Success: %v\n", resp.Success)
	fmt.Printf("  Message: %s\n", resp.Message)
	fmt.Printf("  MessageID: %s\n", resp.MessageId)
}
