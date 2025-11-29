package models

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func GetConnectionToRedis() (*redis.Client, error) {
	var err error

	if redisClient == nil {
		rdb := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password
			DB:       0,  // use default DB
			Protocol: 2,
		})

		ctx := context.Background()
		err := rdb.Ping(ctx).Err()
		if err != nil {
			return nil, err
		}

		redisClient = rdb
	}

	return redisClient, err
}
