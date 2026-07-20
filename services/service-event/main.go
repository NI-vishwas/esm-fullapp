package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"ems-platform/services/service-event/db"
	"ems-platform/services/service-event/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	// By default, it looks for a file named ".env" in the current directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Fetch connection string (with a fallback for running locally)
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// 2. Initialize Database Connection
	repo, err := db.NewMongoRepository(ctx, mongoURI, "event_platform")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Successfully connected to MongoDB container.")

	// 3. Define HTTP handler to serve the catalog
	router := gin.Default()

	// GET /events - Fetch all events
	router.GET("/events", func(c *gin.Context) {
		events, err := repo.GetAllEvents(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
			return
		}
		c.JSON(http.StatusOK, events)
	})

	// POST /events - Add a new event
	router.POST("/events", func(c *gin.Context) {
		var newEvent models.Event
		if err := c.ShouldBindJSON(&newEvent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		if err := repo.CreateEvent(c.Request.Context(), newEvent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add event"})
			return
		}

		c.JSON(http.StatusCreated, newEvent)
	})

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Event Service running on port %s", port)
	log.Fatal(router.Run(":" + port))
}
