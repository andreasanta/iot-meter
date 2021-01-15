package http

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Metrics(c *gin.Context) {

	log.Println("Received metrics hit")

}