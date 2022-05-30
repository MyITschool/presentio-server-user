package v0

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"presentio-server-user/src/v0/handlers"
	"presentio-server-user/src/v0/repo"
)

type Config struct {
	Db *gorm.DB
}

func SetupRouter(group *gin.RouterGroup, config *Config) {
	userRepo := repo.CreateUserRepo(config.Db)
	followRepo := repo.CreateFollowRepo(config.Db)

	handlers.SetupAuthHandler(group.Group("/auth"), &handlers.AuthHandler{
		UserRepo: userRepo,
	})

	handlers.SetupUserHandler(group.Group("/user"), &handlers.UserHandler{
		UserRepo: userRepo,
	})

	handlers.SetupFollowHandler(group.Group("/follow"), &handlers.FollowHandler{
		FollowRepo: followRepo,
	})
}
