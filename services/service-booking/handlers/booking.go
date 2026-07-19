package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

// 🔌 DECOUPLED INTERFACES
type Locker interface {
    Lock(ctx context.Context, resourceKey string, ttl time.Duration) (string, error)
    Unlock(ctx context.Context, resourceKey string, token string) error
}

// Relational database operations (MySQL)
type BookingRepository interface {
    SaveBooking(ctx context.Context, eventID, userID string, seats int) (string, error)
}

// Object storage operations (S3)
type DocumentStorage interface {
    UploadTicketPDF(ctx context.Context, bookingID string, pdfBytes []byte) (string, error)
}

// 📦 HANDLER CORE
type BookingRequest struct {
    EventID string `json:"eventId"`
    UserID  string `json:"userId"`
    Seats   int    `json:"seats"`
}

type BookingHandler struct {
    DB      BookingRepository // MySQL
    S3      DocumentStorage    // S3
    Locker  Locker
}

// Pass both storage dependencies into the constructor
func NewBookingHandler(db BookingRepository, s3 DocumentStorage, locker Locker) *BookingHandler {
    return &BookingHandler{
        DB:     db,
        S3:     s3,
        Locker: locker,
    }
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
    var req BookingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
        return
    }

    if req.EventID == "" || req.UserID == "" || req.Seats <= 0 {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
        return
    }

    // Acquire lock
    lockKey := "event:" + req.EventID + ":allocation"
    token, err := h.Locker.Lock(r.Context(), lockKey, 5*time.Second)
    if err != nil {
        w.WriteHeader(http.StatusConflict)
        json.NewEncoder(w).Encode(map[string]string{"error": "High traffic, try again"})
        return
    }
    defer h.Locker.Unlock(r.Context(), lockKey, token)

    // 1. Save metadata to MySQL
    bookingID, err := h.DB.SaveBooking(r.Context(), req.EventID, req.UserID, req.Seats)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
        return
    }

    // 2. OPTIONAL/NEXT STEP: Generate PDF bytes here...
    // dummyPdfBytes := []byte("%PDF-1.4 ...") 
    // pdfURL, err := h.S3.UploadTicketPDF(r.Context(), bookingID, dummyPdfBytes)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "id":      bookingID,
        "status":  "CONFIRMED",
        "eventId": req.EventID,
        // "ticketUrl": pdfURL, // You can return this to the client once PDF generation is wired up!
    })
}