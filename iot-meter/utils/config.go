package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var MandatoryEnvVariables = [...]string{
	"SERVER_PORT",
	"SSL_CERT_PATH",
	"SSL_KEY_PATH",
	"JWT_SECRET",
}

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return err
	}

	// Check list of mandatory env variables to be set
	for _, v := range MandatoryEnvVariables {
		if os.Getenv(v) == "" {
			return fmt.Errorf("ENV: %s environvment variable must be set or placed in .env", v)
		}
	}

	return nil
}
