package v0

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
	"os"
	"presentio-server-user/src/v0/models"
)

type Config struct {
	Db *gorm.DB
}

func SetupRouter(group *gin.RouterGroup, config *Config) {
	group.POST("/user/register", registerHandler(config))
}

type RegisterParams struct {
	Token string `json:"token"`
}

func registerHandler(config *Config) func(*gin.Context) {
	return func(c *gin.Context) {
		var params RegisterParams

		err := c.ShouldBindJSON(&params)

		if err != nil {
			c.Status(400)
			return
		}

		payload, err := idtoken.Validate(context.Background(), params.Token, os.Getenv("GOOGLE_API_CLIENT_ID"))

		if err != nil {
			c.Status(401)
			return
		}

		email := fmt.Sprint(payload.Claims["email"])

		db := config.Db

		var user models.User

		result := db.Where("email = ?", email).First(&user)

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.Status(500)
				return
			}
		} else {
			c.Status(403)
			return
		}

		firstName := fmt.Sprint(payload.Claims["given_name"])
		lastName := fmt.Sprint(payload.Claims["family_name"])
		pfp := fmt.Sprint(payload.Claims["picture"])

		user.Email = email
		user.FirstName = firstName
		user.LastName = lastName
		user.PFPUrl = pfp

		result = db.Create(&user)

		if result.Error != nil {
			c.Status(500)
			return
		}

		token, err := createNewToken(user.ID)

		if err != nil {
			fmt.Print(err.Error())
			c.Status(500)
			return
		}

		fmt.Println("Created user")

		c.JSON(200, gin.H{
			"token": token,
		})
	}
}
