package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"ems-platform/services/service-event/db"
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
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		events, err := repo.GetAllEvents(r.Context())
		if err != nil {
			http.Error(w, `{"error": "Failed to fetch events"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(events)
	})

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Event Service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}