package repository

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	repository.Repository[entity.NotificationLog]
	Log *logrus.Logger
}

func NewNotificationRepository(log *logrus.Logger) *NotificationRepository {
	return &NotificationRepository{Log: log}
}

// ExistsForPeriod reports whether a log already exists for the
// (customer_id, product_id, period) dedup key.
func (r *NotificationRepository) ExistsForPeriod(db *gorm.DB, customerID int, productID string, period string) (bool, error) {
	var total int64
	err := db.Model(&entity.NotificationLog{}).
		Where("customer_id = ? AND product_id = ? AND period = ?", customerID, productID, period).
		Count(&total).Error
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *NotificationRepository) ListByStoreID(db *gorm.DB, storeID string) ([]entity.NotificationLog, error) {
	var logs []entity.NotificationLog
	err := db.Where("store_id = ?", storeID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}
