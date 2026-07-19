package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Slug        string             `bson:"slug" json:"slug"`
	Status      string             `bson:"status" json:"status"` // e.g. active, inactive, canceled, 
	Date        time.Time          `bson:"date" json:"date"`
	Venue       VenueInfo          `bson:"venue" json:"venue"`
	Inventory   InventoryInfo      `bson:"inventory" json:"inventory"`
	BannerURL   string             `bson:"banner_url" json:"bannerUrl"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}

type VenueInfo struct {
	Name    string `bson:"name" json:"name"`
	Address string `bson:"address" json:"address"`
}

type InventoryInfo struct {
	TotalSlots     int `bson:"total_slots" json:"totalSlots"`
	AvailableSlots int `bson:"available_slots" json:"availableSlots"`
}