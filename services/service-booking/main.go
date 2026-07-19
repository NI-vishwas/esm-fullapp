package main

import (
	"log"
	"net/http"
	"os"

	"ems-platform/services/service-booking/cache"
	"ems-platform/services/service-booking/database"
	"ems-platform/services/service-booking/features" // ◄ Import features pkg
	"ems-platform/services/service-booking/handlers"
	"ems-platform/services/service-booking/storage"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load the .env file
	// By default, it looks for a file named ".env" in the current directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

    // 1. Init mysql database via GORM
    gormDB, err := database.InitDatabase() // Ensure InitDatabase returns (*gorm.DB, error)
    if err != nil {
        log.Fatalf("Database boot failure: %v", err)
    }
    mysqlRepo := database.NewMySQLRepository(gormDB)

	// 2. Init Redis Client connection
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		log.Fatal("REDIS_PASSWORD not set in environment")
	}
    rClient := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: redisPassword,
        DB:       0,
    })
    locker := cache.NewRedisLocker(rClient)

    // 3. Init Flagsmith Feature Flags Engine
    features.InitFlagsmith()

    // 4. Init Storage Manager (S3)
    s3Store, err := storage.NewS3Storage()
    if err != nil {
        log.Fatalf("Storage boot failure: %v", err)
    }

    // 5. Inject BOTH components into Handlers
    bookingHandler := handlers.NewBookingHandler(mysqlRepo, s3Store, locker)

    http.HandleFunc("/book", bookingHandler.CreateBooking)
    // ... start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Booking Service running securely on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}