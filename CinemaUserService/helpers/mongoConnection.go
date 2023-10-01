package helpers

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var (
	mongoClient      *mongo.Client
	UserCollection   *mongo.Collection
	TicketCollection *mongo.Collection
)

func ConnectToMongoDB() *mongo.Database {
	log.Println("MongoDB connection is Initializing...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://abdurrahmangll66:abdurrahman123@test.lx9jrtf.mongodb.net/"))
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	mongoClient = client
	UserCollection = mongoClient.Database("CinemaDB").Collection("users")
	TicketCollection = mongoClient.Database("CinemaDB").Collection("tickets")
	return client.Database("CinemaDB")
}
