package main

import (

	"iot_meter/db"
	"iot_meter/utils"
	"iot_meter/webhandlers"

	"log"

	"gorm.io/gorm"
)

func runServer(DB *gorm.DB) (error) {

	var server webhandlers.MainServer
	err := server.Initialize(DB)
	if err != nil {
		return err
	}

	server.Run()
	return nil
}


func main() {

	// Load env config variables
	err := utils.LoadConfig()
	if err != nil {
		log.Fatal("Unable to load environment configuration")
	}

	// Load accounts database
	DB, err := db.LoadDatabase()
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Unable to open or migrate database")
	}
	defer sqlDB.Close()

	// Run http server
	err = runServer(DB)
	if err != nil {
		log.Fatal(err.Error())
	}
}
