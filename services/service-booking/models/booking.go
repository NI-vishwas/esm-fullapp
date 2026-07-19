package models

import (
	"time"

	"gorm.io/gorm"
)

// Booking represents a customer's transaction for a specific event.
type Booking struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	EventID      string         `gorm:"index;not null;type:varchar(64)" json:"eventId"`
	UserID       string         `gorm:"not null;type:varchar(64)" json:"userId"`
	Seats        int            `gorm:"not null;type:integer;default:1" json:"seatsCount"`
	Status       string         `gorm:"type:varchar(32);default:'CONFIRMED'" json:"status"` // e.g. PENDING, CONFIRMED, CANCELLED
	TicketPdfURL string         `gorm:"type:text" json:"ticketPdfUrl"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // Enables soft-deletes out of the box
}