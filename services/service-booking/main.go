package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ems-platform/services/service-booking/cache"
	"ems-platform/services/service-booking/flags"
)

// BookingRequest defines the payload sent by the React/React Native clients
type BookingRequest struct {
	EventID string `json:"eventId"`
	UserID  string `json:"userId"`
	Seats   int    `json:"seats"`
}

// Booking represents the document schema stored in MongoDB
type Booking struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventID       primitive.ObjectID `bson:"event_id" json:"eventId"`
	UserID        string             `bson:"user_id" json:"userId"`
	SeatsCount    int                `bson:"seats_count" json:"seatsCount"`
	Status        string             `bson:"status" json:"status"`
	TicketPdfURL  string             `bson:"ticket_pdf_url" json:"ticketPdfUrl"`
	CreatedAt     time.Time          `bson:"created_at" json:"createdAt"`
}

type BookingServer struct {
	MongoClient  *mongo.Client
	CacheService *cache.CacheService
	FlagClient   *flags.FlagsmithClient
}

func main() {
	// 1. Gather configs from environment variables
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	flagsmithKey := getEnv("FLAGSMITH_SERVER_KEY", "sb.mock_key_for_dev_purposes")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Initialize Database Client
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("❌ MongoDB connection failed: %v", err)
	}
	log.Println("🔌 Connected to MongoDB.")

	// 3. Initialize Redis Cache Service
	cacheService, err := cache.NewCacheService(redisAddr, redisPassword, 0)
	if err != nil {
		log.Fatalf("❌ Redis connection failed: %v", err)
	}
	log.Println("⚡ Connected to Redis.")

	// 4. Initialize Flagsmith
	flagClient := flags.NewFlagsmithClient(flagsmithKey)
	log.Println("🚩 Flagsmith client initialized.")

	server := &BookingServer{
		MongoClient:  mongoClient,
		CacheService: cacheService,
		FlagClient:   flagClient,
	}

	// 5. Start Server Routes
	http.HandleFunc("/book", server.HandleBooking)

	port := getEnv("PORT", "8082")
	log.Printf("🚀 Booking Service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// HandleBooking processes a client request safely using a Redis Lock and atomic MongoDB updates
func (bs *BookingServer) HandleBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// 1. Check Flagsmith dynamic control: Is booking open?
	if !bs.FlagClient.IsFeatureEnabled("enable_ticket_bookings") {
		http.Error(w, `{"error": "Ticket booking window is temporarily closed."}`, http.StatusForbidden)
		return
	}

	// Decode incoming request payload
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
		return
	}

	eventObjID, err := primitive.ObjectIDFromHex(req.EventID)
	if err != nil {
		http.Error(w, `{"error": "Invalid Event ID format"}`, http.StatusBadRequest)
		return
	}

	// 2. Acquire Redis Lock to protect concurrency
	// We hold a lock per event seat pool for up to 15 seconds to finish this transaction
	seatLockKey := fmt.Sprintf("%s_inventory", req.EventID)
	locked, err := bs.CacheService.AcquireTicketLock(ctx, req.EventID, seatLockKey, req.UserID, 15*time.Second)
	if err != nil || !locked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "High booking volume detected. Please try reserving again in a moment."}`))
		return
	}
	// Always release our Redis lock at the end of execution to open processing back up
	defer bs.CacheService.ReleaseTicketLock(ctx, req.EventID, seatLockKey)

	// 3. Atomically decrement available slots in MongoDB
	eventCollection := bs.MongoClient.Database("event_platform").Collection("events")
	
	// Ensure we only update if the event is active AND there are enough seats left
	filter := bson.M{
		"_id":                       eventObjID,
		"status":                    "active",
		"inventory.available_slots": bson.M{"$gte": req.Seats},
	}
	update := bson.M{
		"$inc": bson.M{"inventory.available_slots": -req.Seats},
	}

	result, err := eventCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, `{"error": "Database error processing reservation"}`, http.StatusInternalServerError)
		return
	}

	// If no documents matched the filter, there are not enough tickets available
	if result.ModifiedCount == 0 {
		http.Error(w, `{"error": "Requested quantity exceeds available ticket supply."}`, http.StatusBadRequest)
		return
	}

	// 4. Create and persist the Booking record in MongoDB
	bookingCollection := bs.MongoClient.Database("event_platform").Collection("bookings")
	newBooking := Booking{
		ID:           primitive.NewObjectID(),
		EventID:      eventObjID,
		UserID:       req.UserID,
		SeatsCount:   req.Seats,
		Status:       "confirmed",
		TicketPdfURL: "", // Will be filled asynchronously by our S3 worker
		CreatedAt:    time.Now(),
	}

	_, err = bookingCollection.InsertOne(ctx, newBooking)
	if err != nil {
		// Roll back the seat count decrement if the booking document creation fails
		rollbackUpdate := bson.M{"$inc": bson.M{"inventory.available_slots": req.Seats}}
		eventCollection.UpdateOne(ctx, bson.M{"_id": eventObjID}, rollbackUpdate)
		
		http.Error(w, `{"error": "Failed to record booking details"}`, http.StatusInternalServerError)
		return
	}

	// 5. Return success payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newBooking)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}