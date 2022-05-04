package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
	"presentio-server-user/src/v0/repo"
	"presentio-server-user/src/v0/util"
	"strconv"
)

type UserHandler struct {
	UserRepo repo.UserRepo
}

func CreateUserHandler(group *gin.RouterGroup, userRepo repo.UserRepo) {
	handler := UserHandler{
		UserRepo: userRepo,
	}

	group.GET("/info/:id", handler.getInfo)
	group.GET("/info/self", handler.getInfoSelf)
}

func (h *UserHandler) getInfo(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.Status(404)
		return
	}

	h.doGetInfo(userId, c)
}

func (h *UserHandler) getInfoSelf(c *gin.Context) {
	h.doGetInfo(-1, c)
}

func (h *UserHandler) doGetInfo(userId int64, c *gin.Context) {
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

	if userId == -1 {
		userId = claims.ID
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
	c.Header("Vary", "Authorization")
	c.Header("Pragma", "")
	c.Header("Expires", "")
}
