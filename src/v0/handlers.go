package v0

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
	"os"
	"presentio-server-user/src/v0/models"
	"strconv"
)

type Config struct {
	Db *gorm.DB
}

var apiKey = os.Getenv("GOOGLE_API_CLIENT_ID")

func SetupRouter(group *gin.RouterGroup, config *Config) {
	group.POST("/register", registerHandler(config))
	group.POST("/authorize", authorizeHandler(config))
	group.GET("/info/:id", userInfoHandler(config))
	group.GET("/refresh", refreshHandler(config))
}

type UserParams struct {
	Token string `json:"token"`
}

func registerHandler(config *Config) func(*gin.Context) {
	return func(c *gin.Context) {
		var params UserParams

		err := c.ShouldBindJSON(&params)

		if err != nil {
			c.Status(400)
			return
		}

		payload, err := idtoken.Validate(context.Background(), params.Token, apiKey)

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

		accessToken, err := createNewAccessToken(user.ID)

		if err != nil {
			c.Status(500)
			return
		}

		refreshToken, err := createNewRefreshToken(user.ID)

		if err != nil {
			c.Status(500)
			return
		}

		c.JSON(200, gin.H{
			"accessToken":   accessToken,
			"refresh_token": refreshToken,
		})
	}
}

func authorizeHandler(config *Config) func(*gin.Context) {
	return func(c *gin.Context) {
		var params UserParams

		err := c.ShouldBindJSON(params)

		if err != nil {
			c.Status(400)
			return
		}

		payload, err := idtoken.Validate(context.Background(), params.Token, apiKey)

		if err != nil {
			c.Status(401)
			return
		}

		email := fmt.Sprint(payload.Claims["email"])

		db := config.Db

		var user models.User

		result := db.Where("email = ?", email).First(&user)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.Status(403)
			} else {
				c.Status(500)
			}

			return
		}

		accessToken, err := createNewAccessToken(user.ID)

		if err != nil {
			c.Status(500)
			return
		}

		refreshToken, err := createNewRefreshToken(user.ID)

		if err != nil {
			c.Status(500)
			return
		}

		c.JSON(200, gin.H{
			"accessToken":   accessToken,
			"refresh_token": refreshToken,
		})
	}
}

func userInfoHandler(config *Config) func(*gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		token, err := validateAccessTokenHeader(authHeader)

		if err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				c.Status(406)
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				c.Status(408)
			} else {
				c.Status(400)
			}

			return
		}

		claims, ok := token.Claims.(*UserClaims)

		if !ok {
			c.Status(403)
			return
		}

		userId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			c.Status(404)
			return
		}

		var user models.User

		db := config.Db

		result := db.Where("id = ?", userId).First(&user)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.Status(404)
			} else {
				c.Status(500)
			}

			return
		}

		c.JSON(200, gin.H{
			"self": user.ID == claims.ID,
			"user": user,
		})
	}
}

func refreshHandler(config *Config) func(*gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		token, err := validateRefreshTokenHeader(authHeader)

		if err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				c.Status(406)
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				c.Status(408)
			} else {
				c.Status(400)
			}

			return
		}

		claims, ok := token.Claims.(*UserClaims)

		if !ok {
			c.Status(403)
			return
		}

		accessToken, err := createNewAccessToken(claims.ID)

		if err != nil {
			c.Status(500)
			return
		}

		c.JSON(200, gin.H{
			"token": accessToken,
		})
	}
}
