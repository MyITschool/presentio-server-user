package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"presentio-server-user/src/v0/models"
	"presentio-server-user/src/v0/repo"
	"presentio-server-user/src/v0/util"
	"strconv"
)

type FollowHandler struct {
	FollowRepo repo.FollowRepo
}

func SetupFollowHandler(group *gin.RouterGroup, handler *FollowHandler) {
	group.POST("/:id", handler.followUser)
}

func (h *FollowHandler) followUser(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.Status(404)
		return
	}

	token, err := util.ValidateAccessTokenHeader(c.GetHeader("Authorization"))

	if err != nil {
		c.Status(util.HandleTokenError(err))
		return
	}

	claims, ok := token.Claims.(*util.UserClaims)

	if !ok || claims.ID == userId {
		c.Status(403)
		return
	}

	err = h.FollowRepo.Transaction(func(tx *gorm.DB) error {
		userRepo := repo.CreateUserRepo(tx)
		followRepo := repo.CreateFollowRepo(tx)

		_, err := followRepo.FindByIds(claims.ID, userId)

		if err == nil {
			c.Status(409)
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		rows, err := userRepo.IncrementFollowers(userId)

		if rows == 0 {
			c.Status(404)
			return nil
		}

		if err != nil {
			return err
		}

		err = followRepo.Create(&models.Follow{
			FromID: claims.ID,
			ToID:   userId,
		})

		if err != nil {
			return err
		}

		c.Status(201)
		return nil
	})

	if err != nil {
		c.Status(500)
		return
	}
}

func (h *FollowHandler) unfollowUser(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.Status(404)
		return
	}

	token, err := util.ValidateAccessTokenHeader(c.GetHeader("Authorization"))

	if err != nil {
		c.Status(util.HandleTokenError(err))
		return
	}

	claims, ok := token.Claims.(*util.UserClaims)

	if !ok || claims.ID == userId {
		c.Status(403)
		return
	}

	err = h.FollowRepo.Transaction(func(tx *gorm.DB) error {
		userRepo := repo.CreateUserRepo(tx)
		followRepo := repo.CreateFollowRepo(tx)

		_, err := followRepo.FindByIds(claims.ID, userId)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(409)
			return nil
		}

		if err != nil {
			return err
		}

		rows, err := userRepo.DecrementFollowers(userId)

		if rows == 0 {
			c.Status(404)
			return nil
		}

		if err != nil {
			return err
		}

		_, err = followRepo.Delete(&models.Follow{
			FromID: claims.ID,
			ToID:   userId,
		})

		if err != nil {
			return err
		}

		c.Status(204)
		return nil
	})

	if err != nil {
		c.Status(500)
		return
	}
}
