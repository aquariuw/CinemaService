package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventType struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name"`
	Price int                `json:"price" bson:"price"`
}
