package repository

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StoreRepository struct {
	repository.Repository[entity.Store]
	Log *logrus.Logger
}

func NewStoreRepository(log *logrus.Logger) *StoreRepository {
	return &StoreRepository{Log: log}
}

func (r *StoreRepository) FindByIdAndUserID(db *gorm.DB, store *entity.Store, id string, userID string) error {
	return db.Where("id = ? AND user_id = ?", id, userID).Take(store).Error
}

func (r *StoreRepository) FindByUserID(db *gorm.DB, store *entity.Store, userID string) error {
	return db.Where("user_id = ?", userID).Take(store).Error
}

func (r *StoreRepository) ListAll(db *gorm.DB) ([]entity.Store, error) {
	var stores []entity.Store
	if err := db.Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}
