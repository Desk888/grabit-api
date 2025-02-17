package initializers

import (
	"fmt"
	"log"
	"context"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client // Initialize the Redis client

func InitRedis()  {
	// Create a new Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Redis host (Docker)
		Password: "", // No password set
		DB:       0, // Utilise default DB
	})

	// Test the connection
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err) // remove in prod
		log.Println("Error connecting to Redis:", err)
	}
	log.Println("Connected to Redis successfully")
}

