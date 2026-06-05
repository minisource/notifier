package main

import (
	"context"
	"fmt"
	"os"

	"github.com/minisource/go-sdk/auth"
	notifier "github.com/minisource/go-sdk/notifier"
)

func main() {
	ac := auth.NewClient(auth.ClientConfig{
		BaseURL:      "http://127.0.0.1:9001",
		ClientID:     "auth-service",
		ClientSecret: "auth-service-secret-key",
	})
	nc, err := notifier.NewClient(context.Background(), notifier.Config{
		Address:    "127.0.0.1:9003",
		AuthClient: ac,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "new client:", err)
		os.Exit(1)
	}
	id, err := nc.SendSMSWithData(context.Background(), &notifier.SMSRequest{
		Phone:    "+989126581160",
		Template: "verify",
		Data:     map[string]string{"code": "123456"},
	})
	fmt.Println("id=", id, "err=", err)
}
