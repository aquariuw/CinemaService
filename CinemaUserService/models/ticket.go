package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Ticket struct {
	Username string             `json:"username"`
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Film     string             `json:"film"`
	Date     time.Time          `json:"date"`
	Price    int                `json:"price"`
	Quantity int                `json:"quantity"`
}
