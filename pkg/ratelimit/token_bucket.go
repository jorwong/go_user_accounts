package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jorwong/go_user_accounts/models"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
	"time"
)

const (
	RATE_LIMIT              = 10
	REFILL_RATE_PER_MINUITE = 1
)

func IsAllowed(email string) bool {
	// 🔑 Keys for Redis storage
	keyLastRefilled := "rate_limit:" + email + ":last_refilled"
	keyTokenCount := "rate_limit:" + email + ":token_count"

	currentTime := time.Now()

	rdb, err := models.GetConnectionToRedis()
	if err != nil {
		fmt.Println("Error getting Redis connection:", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Variable to hold the retrieved and calculated values
	var lastTokenCount int64
	var allowed bool

	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		strTime, err := tx.Get(ctx, keyLastRefilled).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		retrievedCountStr, err := tx.Get(ctx, keyTokenCount).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		if retrievedCountStr == "" {
			lastTokenCount = RATE_LIMIT
		} else {
			lastTokenCount, err = strconv.ParseInt(retrievedCountStr, 10, 64)
			if err != nil {
				return err
			}
		}

		var lastRetrievedTime time.Time
		if strTime == "" {
			lastRetrievedTime = currentTime
		} else {
			lastRetrievedTime, err = time.Parse(time.RFC3339, strTime)
			if err != nil {
				return err
			}
		}

		timePassedMins := currentTime.Sub(lastRetrievedTime).Minutes()
		rawTokenToRefill := timePassedMins * float64(REFILL_RATE_PER_MINUITE)
		tokenToRefill := int64(rawTokenToRefill)

		fmt.Println("tokenToRefill", tokenToRefill)
		// Clamp the new token count at the maximum bucket size
		tokenCount := int(math.Min(float64(lastTokenCount+tokenToRefill), RATE_LIMIT))

		fmt.Println("token count: ", tokenCount)
		fmt.Println("lastTokenCount: ", lastTokenCount)
		// Check if request is allowed
		allowed = false
		if tokenCount >= 1 {
			allowed = true
			tokenCount -= 1 // Consume one token
		}

		// 3. Queue the SET commands for the new state
		// Use time.RFC3339 format for consistent storage/retrieval
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, keyLastRefilled, currentTime, 0)
			pipe.Set(ctx, keyTokenCount, tokenCount, 0)
			return nil
		})

		if allowed {
			return nil
		}

		return errors.New("RATE_LIMITED")
	}, keyLastRefilled, keyTokenCount)

	// 4. Check the result of the TxPipelined
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// key doesnt exist
			return true
		}
		if err.Error() == "RATE_LIMITED" {
			// This is the expected case when tokenCount was < 1
			return false
		}
		fmt.Println("Error in Redis transaction pipeline:", err)
		return false // Network error or other Redis error
	}

	return true // Allowed and updates were made atomically
}
