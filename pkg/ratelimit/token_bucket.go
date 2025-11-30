package pkg

import (
	"context"
	"fmt"
	"github.com/jorwong/go_user_accounts/models"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	RATE_LIMIT              = 10
	REFILL_RATE_PER_MINUITE = 1
)

var tokenBucketScript = redis.NewScript(`
	local token_key = KEYS[1]
	local timestamp_key = KEYS[2]
	
	local rate_limit = tonumber(ARGV[1])
	local refill_rate = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])
	
	local fill_time = 60 / refill_rate
	local ttl = math.floor(fill_time / refill_rate)

-- Get current state
	local last_refilled = tonumber(redis.call("get", timestamp_key))
	local token_count = tonumber(redis.call("get", token_key))

-- initalise if missing
	if not last_refilled then
		last_refilled = now
		token_count = rate_limit
	end

-- calculate refill
	local time_passed = math.max(0, now - last_refilled)
	local tokens_to_add = math.floor(time_passed * (refill_rate / 60))
	
	if tokens_to_add > 0 then
		token_count = math.min(token_count + tokens_to_add, rate_limit)
		last_refilled = now -- Update timestamp only when we added tokens to prevent drift
	end

	local allowed = 0
	if token_count >= requested then
		token_count = token_count - requested
		allowed = 1
		-- save new state
		redis.call("set", token_key, token_count)
		redis.call("set", timestamp_key, last_refilled)
	end
	
	return allowed
`)

func IsAllowed(email string) (bool, error) {
	//only the email is hashed next time for sharding
	keyTokenCount := "rate_limit:{" + email + "}:token_count"
	keyLastRefilled := "rate_limit:{" + email + "}:last_refilled"

	currentTime := time.Now().Unix()

	rdb := models.GetConnectionToRedis()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := tokenBucketScript.Run(ctx, rdb, []string{keyTokenCount, keyLastRefilled}, RATE_LIMIT, REFILL_RATE_PER_MINUITE, currentTime, 1).Result()
	if err != nil {
		fmt.Println("Error executing LUA script:", err)
		return false, err
	}

	return result.(int64) == 1, nil
}
