package database

import (
    "fmt"
    "log"
    "os"
    "time"

    "ems-platform/services/service-booking/models"

    "gorm.io/driver/mysql" // Swapped from postgres
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// DB represents our global thread-safe database connection handle.
var DB *gorm.DB

// InitDatabase establishes a MySQL connection and migrates schemas.
func InitDatabase() (*gorm.DB, error) {
    // Retrieve connection string from the environment, falling back to a local dev db.
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // Updated fallback string to MySQL DSN format with time parsing enabled
        dsn = "root:root@tcp(localhost:3306)/booking_db?charset=utf8mb4&parseTime=True&loc=Local"
    }

    var db *gorm.DB
    var err error

    // Retry connection loop. Docker container databases can take a few seconds to accept connections.
    maxRetries := 5
    for i := 1; i <= maxRetries; i++ {
        log.Printf("Connecting to MySQL database (Attempt %d/%d)...", i, maxRetries) // Updated log text to MySQL
        
        db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{ // Swapped to mysql.Open
            Logger: logger.Default.LogMode(logger.Info), // Log queries to console in development
        })
        
        if err == nil {
            break
        }

        log.Printf("Failed to connect: %v. Retrying in 3 seconds...", err)
        time.Sleep(3 * time.Second)
    }

    if err != nil {
        return nil, fmt.Errorf("could not connect to MySQL after %d attempts: %w", maxRetries, err) // Updated error text to MySQL
    }

    // Configure the underlying connection pool parameters
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB handle: %w", err)
    }

    sqlDB.SetMaxIdleConns(10)           // Keep up to 10 idle connections open
    sqlDB.SetMaxOpenConns(100)          // Limit peak simultaneous connections to 100
    sqlDB.SetConnMaxLifetime(time.Hour) // Close connections older than 1 hour to recycle resources

    log.Println("Successfully connected to MySQL database.") // Updated log text to MySQL

    // Execute Auto-Migration
    log.Println("Running database schema auto-migrations...")
    err = db.AutoMigrate(&models.Booking{})
    if err != nil {
        return nil, fmt.Errorf("auto-migration failed: %w", err)
    }
    
    log.Println("Database auto-migration completed successfully.")

    // Set global DB handle
    DB = db
    return DB, nil
}