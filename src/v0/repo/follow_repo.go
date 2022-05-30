package repo

import (
	"database/sql"
	"gorm.io/gorm"
	"presentio-server-user/src/v0/models"
)

type FollowRepo struct {
	db *gorm.DB
}

func CreateFollowRepo(db *gorm.DB) FollowRepo {
	return FollowRepo{
		db: db,
	}
}

func (r *FollowRepo) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return r.db.Transaction(fc, opts...)
}

func (r *FollowRepo) FindByIds(from int64, to int64) (*models.Follow, error) {
	var follow models.Follow

	result := r.db.
		Where("from_id = ?", from).
		Where("to_id = ?", to).
		First(&follow)

	return &follow, result.Error
}

func (r *FollowRepo) Create(follow *models.Follow) error {
	return r.db.Create(follow).Error
}

func (r *FollowRepo) Delete(follow *models.Follow) (int64, error) {
	tx := r.db.Delete(follow)

	return tx.RowsAffected, tx.Error
}
