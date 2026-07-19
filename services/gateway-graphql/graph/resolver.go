package graph

import (
	"context"
	"ems-platform/services/gateway-graphql/graph/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Resolver acts as our dependency injection hub for API services.
type Resolver struct {
	EventServiceURL   string
	BookingServiceURL string
	HTTPClient        *http.Client
}

func NewResolver() *Resolver {
	// Load the .env file
	// By default, it looks for a file named ".env" in the current directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Pull service endpoints from environmental configuration
	eventURL := os.Getenv("EVENT_SERVICE_URL")
	if eventURL == "" {
		eventURL = "http://localhost:8081" // Local fallback
	}

	bookingURL := os.Getenv("BOOKING_SERVICE_URL")
	if bookingURL == "" {
		bookingURL = "http://localhost:8082" // Local fallback
	}

	return &Resolver{
		EventServiceURL:   eventURL,
		BookingServiceURL: bookingURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second, // Prevent hung threads from blocking gateway
		},
	}
}

// // ----------------------------------------------------
// // QUERY RESOLVER: Get Active Events
// // ----------------------------------------------------

func (r *Resolver) Events(ctx context.Context) ([]*model.Event, error) {
	reqUrl := fmt.Sprintf("%s/events", r.EventServiceURL)

	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach event service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("event service responded with status: %d", resp.StatusCode)
	}

	var events []*model.Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode event catalog: %w", err)
	}

	return events, nil
}
