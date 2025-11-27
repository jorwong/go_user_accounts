package models

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "host=localhost user=user password=my_strong_postgres_password dbname=mydatabase port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&User{}, &Session{})
	if err != nil {
		log.Fatalf("Failed to migrate table: %v", err)
	}

	fmt.Println("Connected to database")
}
