package db

import (
	"context"
	// "time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ems-platform/services/service-event/models"
)

type MongoRepository struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoRepository(ctx context.Context, uri, dbName string) (*MongoRepository, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoRepository{
		Client: client,
		DB:     client.Database(dbName),
	}, nil
}

// GetAllEvents retrieves all active events from MongoDB
func (r *MongoRepository) GetAllEvents(ctx context.Context) ([]models.Event, error) {
	collection := r.DB.Collection("events")

	// Filter for active events
	filter := bson.M{"status": "active"}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *MongoRepository) GetEventByID(ctx context.Context, eventID string) (*models.Event, error) {
	collection := r.DB.Collection("events")

	filter := bson.M{"_id": eventID}
	
	var event models.Event
	err := collection.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No event found
		}
		return nil, err
	}

	return &event, nil
}

func (r *MongoRepository) CreateEvent(ctx context.Context, event models.Event) error {
	collection := r.DB.Collection("events")
	
	_, err := collection.InsertOne(ctx, event)
	return err
}

func (r *MongoRepository) UpdateEvent(ctx context.Context, eventID string, update bson.M) error {
	collection := r.DB.Collection("events")
	
	filter := bson.M{"_id": eventID}
	updateDoc := bson.M{"$set": update}
	_, err := collection.UpdateOne(ctx, filter, updateDoc)
	return err
}

