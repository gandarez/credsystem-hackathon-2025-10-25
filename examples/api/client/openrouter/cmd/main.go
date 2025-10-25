package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"openrouter"
)

func main() {
	// Define base URL and API key
	baseURL := "https://openrouter.ai/api/v1"
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENROUTER_API_KEY environment variable is not set.")
		return
	}

	// Create a new OpenRouter client with authentication
	client := openrouter.NewClient(baseURL, openrouter.WithAuth(apiKey))

	// Define context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Define the intent for the request
	intent := "bom dia, quero ter um limite maior"

	// Call the ChatCompletion function
	response, err := client.ChatCompletion(ctx, intent)
	if err != nil {
		fmt.Printf("Error calling ChatCompletion: %v\n", err)
		return
	}

	// Print the response
	fmt.Printf("Service ID: %d\n", response.ServiceID)
	fmt.Printf("Service Name: %s\n", response.ServiceName)
}
