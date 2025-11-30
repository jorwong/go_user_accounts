package models

import (
	"github.com/redis/go-redis/v9"
	"sync"
)

var (
	rdb  *redis.Client
	once sync.Once // 1. Add this variable
)

func GetConnectionToRedis() *redis.Client {
	if rdb != nil {
		return rdb
	}
	once.Do(func() {
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379", // Check your address
			Password: "",               // No password set
			DB:       0,                // Use default DB
		})
	})

	return rdb
}
