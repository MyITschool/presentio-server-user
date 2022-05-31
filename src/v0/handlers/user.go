package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"presentio-server-user/src/v0/repo"
	"presentio-server-user/src/v0/util"
	"strconv"
)

type UserHandler struct {
	UserRepo repo.UserRepo
}

func SetupUserHandler(group *gin.RouterGroup, handler *UserHandler) {
	group.GET("/info/:id", handler.getInfo)
	group.GET("/info/self", handler.getInfoSelf)
	group.GET("/search/:page", handler.search)
}

func (h *UserHandler) search(c *gin.Context) {
	token, err := util.ValidateAccessTokenHeader(c.GetHeader("Authorization"))

	if err != nil {
		c.Status(util.HandleTokenError(err))
		return
	}

	page, err := strconv.Atoi(c.Param("page"))

	if err != nil {
		c.Status(404)

		return
	}

	_, ok := token.Claims.(*util.UserClaims)

	if !ok {
		c.Status(403)
		return
	}

	keywords := c.QueryArray("keyword")

	users, err := h.UserRepo.FindByQuery(keywords, page)

	c.Header("Cache-Control", "public, max-age=300")
	c.Header("Pragma", "")
	c.Header("Expires", "")
	c.JSON(200, users)
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
		c.Status(util.HandleTokenError(err))

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

	user, err := h.UserRepo.FindById(userId, claims.ID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(404)
		} else {
			c.Status(500)
		}

		return
	}

	cache := "public, max-age="

	if userId == -1 {
		cache += "300"
	} else {
		cache += "18000"
	}

	c.Header("Cache-Control", cache)
	c.Header("Pragma", "")
	c.Header("Expires", "")

	c.JSON(200, gin.H{
		"self": user.ID == claims.ID,
		"user": user,
	})
}
