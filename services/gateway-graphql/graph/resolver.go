package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"ems-platform/services/gateway-graphql/graph/model"
)

// Resolver acts as our dependency injection hub for API services.
type Resolver struct {
	EventServiceURL   string
	BookingServiceURL string
	HTTPClient        *http.Client
}

func NewResolver() *Resolver {
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

// ----------------------------------------------------
// QUERY RESOLVER: Get Active Events
// ----------------------------------------------------

func (r *Resolver) Query_GetEvents(ctx context.Context) ([]*model.Event, error) {
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

// ----------------------------------------------------
// MUTATION RESOLVER: Create Booking
// ----------------------------------------------------

// BookingRequestPayload defines the JSON object format expected by the booking microservice
type BookingRequestPayload struct {
	EventID string `json:"eventId"`
	UserID  string `json:"userId"`
	Seats   int    `json:"seats"`
}

func (r *Resolver) Mutation_CreateBooking(ctx context.Context, eventID string, userID string, seats int) (*model.Booking, error) {
	reqUrl := fmt.Sprintf("%s/book", r.BookingServiceURL)

	// Build the transaction request payload
	payload := BookingRequestPayload{
		EventID: eventID,
		UserID:  userID,
		Seats:   seats,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach booking service: %w", err)
	}
	defer resp.Body.Close()

	// Handle downstream business errors gracefully (e.g., locking conflicts or out of stock)
	if resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if msg, exists := errorResponse["error"]; exists {
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("booking service rejected transaction with status: %d", resp.StatusCode)
	}

	var booking model.Booking
	if err := json.NewDecoder(resp.Body).Decode(&booking); err != nil {
		return nil, fmt.Errorf("failed to decode booking details: %w", err)
	}

	return &booking, nil
}