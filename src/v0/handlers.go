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

	handlers.CreateAuthHandler(group.Group("/auth"), userRepo)
}
