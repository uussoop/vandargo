// cmd/test.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/uussoop/vandargo"
)

func main() {
	// Initialize configuration
	config := vandargo.ConfigWrapper{
		vandargo.Config{
			APIKey:      "api_key", // Replace with your actual API key
			BaseURL:     "https://ipg.vandar.io",
			SandboxMode: true,
			Timeout:     30,
			CallbackURL: "https://test.com/callback", // Replace with your actual callback URL
		},
	}

	// Create storage and logger
	storage := vandargo.NewMemoryStorage()
	logger := vandargo.NewSimpleLogger("DEBUG")

	// Create a new Vandar client
	client, err := vandargo.NewClient(&config, storage, logger)
	if err != nil {
		log.Fatalf("Failed to create Vandar client: %v", err)
	}

	// Create a transaction
	transaction := &vandargo.PaymentInitRequest{
		Amount:      1000000, // Amount in Rials
		CallbackURL: config.GetCallbackURL(),
		Description: "Payment for product",
		Mobile:      "09123456789", // Optional
	}

	// Call the payment initialization method
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create metadata for the transaction
	metadata := map[string]string{
		"customer_id": "12345",
		"order_id":    "ORD-98765",
	}

	// Initialize payment
	response, err := client.InitiatePayment(ctx, transaction.Amount, transaction.Description, metadata)
	if err != nil {
		log.Fatalf("Failed to initiate payment: %v", err)
	}

	// Print the response
	fmt.Printf("Payment Token: %s\n", response.Token)
	fmt.Printf("Payment Status: %t\n", response.Status)
	fmt.Printf("Payment Message: %s\n", response.Message)

	// Print payment URL
	fmt.Printf("Payment URL: %s/v4/%s\n", config.GetBaseURL(), response.Token)
	fmt.Println("\nPlease open the above URL in your browser to complete the payment.")
	fmt.Println("After payment, you will be redirected to the callback URL.")

	// Wait for user input to verify payment
	fmt.Println("\nPress Enter after completing the payment to verify it...")
	fmt.Scanln()

	// Get transaction information
	info, err := client.GetTransactionInfo(ctx, "YOUR_PAYMENT_TOKEN")
	if err != nil {
		log.Fatalf("Failed to get transaction info: %v", err)
	}

	// Check the response
	if info.Status == 1 {
		fmt.Printf("Transaction Amount: %s Rials\n", info.Amount)
		fmt.Printf("Transaction ID: %d\n", info.TransID)
		fmt.Printf("Reference Number: %s\n", info.RefNumber)
		fmt.Printf("Tracking Code: %s\n", info.TrackingCode)
		fmt.Printf("Card Number: %s\n", info.CardNumber)
		fmt.Printf("Created At: %s\n", info.CreatedAt)
		fmt.Printf("Payment Date: %s\n", info.PaymentDate)
	} else {
		fmt.Printf("Error Code: %d\n", info.Code)
		fmt.Printf("Error Message: %s\n", info.Message)
	}
	// Verify payment
	verifyResponse, err := client.VerifyPayment(ctx, response.Token)
	if err != nil {
		log.Fatalf("Failed to verify payment: %v", err)
	}

	// Print verification response
	fmt.Printf("\nVerification Status: %t\n", verifyResponse.Status)
	if verifyResponse.Status {
		fmt.Printf("Amount: %d Rials\n", verifyResponse.Amount)
		fmt.Printf("Reference ID: %s\n", verifyResponse.RefID)
		fmt.Printf("Card Number: %s\n", verifyResponse.CardNumber)
	} else {
		fmt.Printf("Verification Message: %s\n", verifyResponse.Message)
	}
}
