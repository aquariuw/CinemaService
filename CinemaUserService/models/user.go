package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Username     string             `json:"username"`
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	LoyaltyPoint int                `json:"loyaltyPoint" bson:"loyaltyPoint"`
	FreeTicket   int                `json:"freeTicket" bson:"freeTicket"`
	Balance      int                `json:"balance" bson:"balance"`
}
