package repo

import (
	"gorm.io/gorm"
	"presentio-server-user/src/v0/models"
	"strings"
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

func (r *UserRepo) FindById(id int64, myUserId int64) (*models.User, error) {
	var user models.User

	result := r.db.
		Where("users.id = ?", id).
		Joins("Follow", r.db.Where(&models.Follow{FromID: myUserId})).
		First(&user)

	return &user, result.Error
}

func (r *UserRepo) Create(user *models.User) error {
	result := r.db.Create(user)

	return result.Error
}

func (r *UserRepo) IncrementFollowers(userId int64) (int64, error) {
	tx := r.db.
		Exec("UPDATE users SET followers = followers + 1 WHERE id = ?", userId)

	return tx.RowsAffected, tx.Error
}

func (r *UserRepo) DecrementFollowers(userId int64) (int64, error) {
	tx := r.db.
		Exec("UPDATE users SET followers = followers - 1 WHERE id = ?", userId)

	return tx.RowsAffected, tx.Error
}

func (r *UserRepo) IncrementFollowing(userId int64) (int64, error) {
	tx := r.db.
		Exec("UPDATE users SET following = following + 1 WHERE id = ?", userId)

	return tx.RowsAffected, tx.Error
}

func (r *UserRepo) DecrementFollowing(userId int64) (int64, error) {
	tx := r.db.
		Exec("UPDATE users SET following = following - 1 WHERE id = ?", userId)

	return tx.RowsAffected, tx.Error
}

func (r *UserRepo) FindByQuery(keywords []string, page int) ([]models.User, error) {
	var results []models.User

	tx := r.db.
		Where("name @@ to_tsquery('english', ?)", strings.Join(keywords, "&")).
		Limit(20).
		Offset(20 * page).
		Find(&results)

	return results, tx.Error
}
