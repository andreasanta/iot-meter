package db

import (
	"log"
	"os"

	"iot_meter/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadDatabase() (*gorm.DB, error) {

	log.Println("Opening database")

	db, err := gorm.Open(sqlite.Open(os.Getenv("DB_FILE_PATH")), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Println("Unable to open SQLLite Database")
		return nil, err
	}

	log.Println("Performing migrations")

	// Perform automatic migrations to update database
	err = db.AutoMigrate(&models.Account{}, &models.User{})
	if err != nil {
		log.Println("Unable to migrate DB automatically:")
		log.Println(err)
		return nil, err
	}

	return db, nil
}
