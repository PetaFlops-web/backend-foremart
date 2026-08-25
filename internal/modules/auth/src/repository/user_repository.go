package repository

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/auth/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository struct {
	repository.Repository[entity.User]
	Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger) *UserRepository {
	return &UserRepository{Log: log}
}

func (r *UserRepository) FindByUsername(db *gorm.DB, user *entity.User, username string) error {
	return db.Where("username = ?", username).Take(user).Error
}

func (r *UserRepository) CountByUsername(db *gorm.DB, username string) (int64, error) {
	var total int64

	err := db.Model(new(entity.User)).
		Where("username = ?", username).
		Count(&total).Error
	return total, err
}