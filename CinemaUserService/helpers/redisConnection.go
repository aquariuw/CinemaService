package helpers

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

func ConnectToRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the redis...")
	return client
}
