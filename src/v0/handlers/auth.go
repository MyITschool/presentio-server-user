package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
	"os"
	"presentio-server-user/src/v0/repo"
	"presentio-server-user/src/v0/util"
	"strconv"
)

type AuthHandler struct {
	UserRepo repo.UserRepo
}

type AuthParams struct {
	Token string `json:"token"`
}

var apiKey = os.Getenv("GOOGLE_API_CLIENT_ID")

func CreateAuthHandler(group *gin.RouterGroup, userRepo repo.UserRepo) {
	handler := AuthHandler{
		UserRepo: userRepo,
	}

	group.POST("/register", handler.register)
	group.POST("/authorize", handler.authorize)
	group.GET("/info/:id", handler.get)
	group.GET("/refresh", handler.refresh)
}

func (h *AuthHandler) register(c *gin.Context) {
	var params AuthParams

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

	user, err := h.UserRepo.FindByEmail(email)

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

	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.PFPUrl = pfp

	err = h.UserRepo.Create(user)

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
		"accessToken":   accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) authorize(c *gin.Context) {
	var params AuthParams

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
		"accessToken":   accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) get(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token, err := util.ValidateAccessTokenHeader(authHeader)

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

	claims, ok := token.Claims.(*util.UserClaims)

	if !ok {
		c.Status(403)
		return
	}

	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.Status(404)
		return
	}

	user, err := h.UserRepo.FindById(userId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	c.Header("Cache-Control", "public, max-age=18000")
	c.Header("Pragma", "")
	c.Header("Expires", "")
}

func (h *AuthHandler) refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token, err := util.ValidateRefreshToken(authHeader)

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
		"token": accessToken,
	})
}
