package models

import (
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

func CreateSession(token string, user *User) (*Session, error) {
	timeExpire := time.Now().Add(time.Duration(time.Second) * 60) // expire after 60 seconds
	newSession := Session{Token: token, UserID: user.ID, Expire: timeExpire}
	result := DB.Create(&newSession)
	if result.Error != nil {
		return nil, result.Error
	}
	return &newSession, nil
}
