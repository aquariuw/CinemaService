package helpers

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func ConnectToRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis sunucusunun adresi ve bağlantı noktası
		Password: "",               // Redis sunucusu için şifre (gerekiyorsa)
		DB:       0,                // Redis veritabanı numarası
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Redis'e bağlandı...")
}
