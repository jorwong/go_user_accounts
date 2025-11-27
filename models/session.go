package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/jorwong/go_user_accounts/connections"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

// Session Model (Assuming a simple session table)
type Session struct {
	gorm.Model
	Token  string `gorm:"unique"`
	Expire time.Time
	UserID uint // Foreign key to User.ID
	User   User // GORM automatically handles the association
}

func CreateSession(user *User, timeExpire time.Time, token string) (*Session, error) {

	newSession := Session{Token: token, UserID: user.ID, Expire: timeExpire}
	result := DB.Create(&newSession)
	if result.Error != nil {
		return nil, result.Error
	}
	return &newSession, nil
}

// userId -> session
func CreateRedisSession(token string, expireTime time.Time, user *User) (string, bool, error) {
	rdb, err := connections.GetConnection()
	if err != nil {
		return "", false, err
	}
	// Create Redis Session if it doesnt exist in the DB else GET
	key := fmt.Sprintf("session:%s", user.Email)
	durationUntilExpiration := expireTime.Sub(time.Now())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if durationUntilExpiration <= 0 {
		return "", false, fmt.Errorf("session expired at %v", durationUntilExpiration)
	}

	cmd := rdb.SetArgs(ctx, key, token, redis.SetArgs{
		TTL:  durationUntilExpiration,
		Get:  true,
		Mode: string(redis.NX),
	})

	existingValue, err := cmd.Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		// redis.Nil error means the key did not exist before the set. This is not a failure.
		return "", false, fmt.Errorf("redis SET command failed for user %s: %w", user.UserID, err)
	}

	if existingValue == "" || errors.Is(err, redis.Nil) {
		return token, true, nil
	}

	return existingValue, false, nil
}

func RevokeSession(user *User) error {
	rdb, err := connections.GetConnection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("session:%s", user.Email)
	err = rdb.Del(ctx, key).Err()

	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

func GetTokenFromRedis(email string) (string, error) {
	rdb, err := connections.GetConnection()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := fmt.Sprintf("session:%s", email)

	sessionKey, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return sessionKey, nil
}
