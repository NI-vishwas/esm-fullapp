package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ems-platform/services/service-booking/database"
	"ems-platform/services/service-booking/models"
	"ems-platform/services/service-booking/storage"
	"ems-platform/services/service-booking/ticket"

	"github.com/google/uuid"
)

// BookingRequest defines the payload structure incoming from our GraphQL gateway
type BookingRequest struct {
	EventID string `json:"eventId"`
	UserID  string `json:"userId"`
	Seats   int    `json:"seats"`
}

// BookingResponse defines what we return back to the gateway on success
type BookingResponse struct {
	ID           string `json:"id"`
	EventID      string `json:"eventId"`
	UserID       string `json:"userId"`
	SeatsCount   int    `json:"seatsCount"`
	Status       string `json:"status"`
	TicketPdfURL string `json:"ticketPdfUrl"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// BookingHandler orchestrates database, PDF generation, and S3 storage
type BookingHandler struct {
	Storage *storage.S3Storage
}

func NewBookingHandler(store *storage.S3Storage) *BookingHandler {
	return &BookingHandler{
		Storage: store,
	}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow POST requests
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 2. Decode incoming payload
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate inputs
	if req.EventID == "" || req.UserID == "" || req.Seats <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "eventId, userId, and a positive seats count are required")
		return
	}

	// Generate a unique ID for this transaction
	bookingID := uuid.New().String()

	// 3. (Mock step) Fetch Event Meta
	// In a real system, you would call service-event over gRPC/HTTP to get the actual event name and price.
	// For this unified workflow, we'll mock the event data.
	eventTitle := "Global Developers Summit 2026"
	eventTime := time.Date(2026, time.September, 15, 9, 0, 0, 0, time.UTC)
	userName := "Developer Attendee" // In reality, pulled from auth context/user service

	// 4. Generate the Ticket PDF in memory
	pdfBytes, err := ticket.GenerateTicketPDF(ticket.TicketData{
		BookingID:   bookingID,
		EventTitle:  eventTitle,
		EventDate:   eventTime,
		UserName:    userName,
		SeatsCount:  req.Seats,
		TicketPrice: 99.00, // Mock price
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate ticket PDF")
		return
	}

	// 5. Upload PDF to S3 Storage
	pdfURL, err := h.Storage.UploadTicketPDF(r.Context(), bookingID, pdfBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to store ticket PDF")
		return
	}

	// 6. Save Booking Record to PostgreSQL via GORM
	booking := models.Booking{
		ID:           bookingID,
		EventID:      req.EventID,
		UserID:       req.UserID,
		SeatsCount:   req.Seats,
		Status:       "CONFIRMED",
		TicketPdfURL: pdfURL,
	}

	// Perform an atomic DB insert
	if err := database.DB.Create(&booking).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Database write failure: "+err.Error())
		return
	}

	// 7. Write JSON Response to Gateway
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(BookingResponse{
		ID:           booking.ID,
		EventID:      booking.EventID,
		UserID:       booking.UserID,
		SeatsCount:   booking.SeatsCount,
		Status:       booking.Status,
		TicketPdfURL: booking.TicketPdfURL,
	})
}

// Helper function to write standardized JSON error structures
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}