package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
	"log"
	"os"
)

type RegisterParams struct {
	Token string `json:"token"`
}

func registerHandler(c *gin.Context) {
	var params RegisterParams

	err := c.ShouldBindJSON(&params)

	if err != nil {
		c.Status(400)
		return
	}

	payload, err := idtoken.Validate(context.Background(), params.Token, os.Getenv("GOOGLE_API_CLIENT_ID"))

	if err != nil {
		c.Status(403)
		return
	}

	c.JSON(200, payload)
}

func main() {
	logger := log.Logger{}

	router := gin.Default()

	v0 := router.Group("/v0")

	v0.POST("/user/register", registerHandler)

	err := router.Run()

	if err != nil {
		logger.Fatalln("Failed to start server on port %s", os.Getenv("PORT"))
	}
}
