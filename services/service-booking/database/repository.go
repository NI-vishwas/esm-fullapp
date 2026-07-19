package database

import (
    "context"
    "fmt"
    "time"
    "gorm.io/gorm"
)

// Define your GORM model matching your MySQL schema
type BookingModel struct {
    ID        string    `gorm:"primaryKey;type:varchar(50)"`
    EventID   string    `gorm:"type:varchar(50);not null"`
    UserID    string    `gorm:"type:varchar(50);not null"`
    Seats     int       `gorm:"not null"`
    CreatedAt time.Time
}

// TableName explicitly matches the DB table
func (BookingModel) TableName() string {
    return "bookings"
}

type MySQLRepository struct {
    db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
    return &MySQLRepository{db: db}
}

// SaveBooking satisfies handlers.BookingRepository
func (r *MySQLRepository) SaveBooking(ctx context.Context, eventID, userID string, seats int) (string, error) {
    bookingID := fmt.Sprintf("bk-%d", time.Now().UnixNano())
    
    newBooking := BookingModel{
        ID:        bookingID,
        EventID:   eventID,
        UserID:    userID,
        Seats:     seats,
        CreatedAt: time.Now(),
    }

    // Insert record into MySQL via GORM
    if err := r.db.WithContext(ctx).Create(&newBooking).Error; err != nil {
        return "", err
    }

    return bookingID, nil
}