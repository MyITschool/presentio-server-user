package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
	"os"
	"presentio-server-user/src/v0/models"
	"presentio-server-user/src/v0/repo"
	"presentio-server-user/src/v0/util"
)

type AuthHandler struct {
	UserRepo repo.UserRepo
}

type AuthParams struct {
	Token string `json:"token"`
}

var apiKey = os.Getenv("GOOGLE_API_CLIENT_ID")

func SetupAuthHandler(group *gin.RouterGroup, handler *AuthHandler) {
	group.POST("/register", handler.register)
	group.POST("/authorize", handler.authorize)
	group.GET("/refresh", handler.refresh)
}

func (h *AuthHandler) register(c *gin.Context) {
	var params AuthParams

	err := c.ShouldBindJSON(&params)

	if err != nil {
		c.Status(422)
		return
	}

	payload, err := idtoken.Validate(context.Background(), params.Token, apiKey)

	if err != nil {
		c.Status(401)
		return
	}

	email := fmt.Sprint(payload.Claims["email"])

	_, err = h.UserRepo.FindByEmail(email)

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
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

	user := models.User{
		Email:  email,
		Name:   firstName + " " + lastName,
		PFPUrl: pfp,
	}

	err = h.UserRepo.Create(&user)

	if err != nil {
		c.Status(500)
		return
	}

	accessToken, err := util.CreateNewAccessToken(user.ID)

	if err != nil {
		c.Status(500)
		return
	}

	refreshToken, err := util.CreateNewRefreshToken(user.ID)

	if err != nil {
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) authorize(c *gin.Context) {
	var params AuthParams

	err := c.ShouldBindJSON(&params)

	if err != nil {
		c.Status(422)
		return
	}

	payload, err := idtoken.Validate(context.Background(), params.Token, apiKey)

	if err != nil {
		c.Status(401)
		return
	}

	email := fmt.Sprint(payload.Claims["email"])

	user, err := h.UserRepo.FindByEmail(email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(403)
		} else {
			c.Status(500)
		}

		return
	}

	accessToken, err := util.CreateNewAccessToken(user.ID)

	if err != nil {
		c.Status(500)
		return
	}

	refreshToken, err := util.CreateNewRefreshToken(user.ID)

	if err != nil {
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token, err := util.ValidateRefreshTokenHeader(authHeader)

	if err != nil {
		c.Status(util.HandleTokenError(err))

		return
	}

	claims, ok := token.Claims.(*util.UserClaims)

	if !ok {
		c.Status(403)
		return
	}

	accessToken, err := util.CreateNewAccessToken(claims.ID)

	if err != nil {
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{
		"accessToken": accessToken,
	})
}
