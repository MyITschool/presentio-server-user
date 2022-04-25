package repo

import (
	"gorm.io/gorm"
	"presentio-server-user/src/v0/models"
)

type UserRepo struct {
	db *gorm.DB
}

func CreateUserRepo(db *gorm.DB) UserRepo {
	return UserRepo{
		db,
	}
}

func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User

	result := r.db.Where("email = ?", email).First(&user)

	return &user, result.Error
}

func (r *UserRepo) FindById(id int64) (*models.User, error) {
	var user models.User

	result := r.db.Where("id = ?", id).First(&user)

	return &user, result.Error
}

func (r *UserRepo) Create(user *models.User) error {
	result := r.db.Create(user)

	return result.Error
}
