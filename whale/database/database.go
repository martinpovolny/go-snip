package database

import (
	"log"
	"whale/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database")
	}

	err = db.AutoMigrate(&models.Person{})
	if err != nil {
		log.Fatal("failed to migrate the database")
	}

	return db
}
