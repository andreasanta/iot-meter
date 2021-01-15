package db

import (
	"log"
	"os"

	"iot_meter/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Global static db
var database *gorm.DB

func LoadDatabase() (*gorm.DB) {

	log.Println("Opening database")

	db, err := gorm.Open(sqlite.Open(os.Getenv("DB_FILE_PATH")), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Performing migrations")

	// Perform automatic migrations to update database
	db.AutoMigrate(&models.Account{}, &models.User{})

	database = db
	return db
}