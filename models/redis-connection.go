package models

import (
	"github.com/redis/go-redis/v9"
)

func GetConnectionToRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Check your address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	return rdb
}
