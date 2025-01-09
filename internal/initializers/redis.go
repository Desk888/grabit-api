package initializers

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Redis host (Docker)
		Password: "", // No password set
		DB:       0, // Utilise default DB
	})

	// Test the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return nil
	}
	fmt.Println("Connected to Redis successfully")
	return client
}

