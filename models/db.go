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

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: false})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&User{}, &Session{})
	if err != nil {
		log.Fatalf("Failed to migrate table: %v", err)
	}

	postgresDB, err := DB.DB()

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	postgresDB.SetMaxIdleConns(30)
	postgresDB.SetMaxOpenConns(100)
	fmt.Println("Connected to database")
}
