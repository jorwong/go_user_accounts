package models

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User Model - Fields must be EXPORTED (start with capital letter)
type User struct {
	gorm.Model      // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	UserID     uint `gorm:"primaryKey;autoIncrement"` // Unique index on UserId
	Name       string
	Email      string `gorm:"unique"` // Unique constraint for Email
	Password   []byte
}

func (u *User) ToString() string {
	return fmt.Sprintf("%s|%s", u.Name, u.Email)
}

func FindUserByEmail(email string, DB *gorm.DB) (*User, error) {
	// search criteria
	searchCriteria := User{
		Email: email,
	}

	var retrievedUser User

	result := DB.Where(&searchCriteria).First(&retrievedUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("USER_NOT_FOUND")
		} else {
			return nil, errors.New("DB_ERROR")
		}
	}

	return &retrievedUser, nil
}

func CreateUser(email string, name string, password string, DB *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	result := DB.Create(&User{Email: email, Name: name, Password: hash})
	return result.Error
}

func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	return err == nil
}
