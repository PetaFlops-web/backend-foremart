package repository

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	repository.Repository[entity.Transaction]
	Log *logrus.Logger
}

func NewTransactionRepository(log *logrus.Logger) *TransactionRepository {
	return &TransactionRepository{Log: log}
}

// FindByIdWithItems finds a transaction by ID and eager-loads its items.
func (r *TransactionRepository) FindByIdWithItems(db *gorm.DB, id string) (*entity.Transaction, error) {
	var txn entity.Transaction
	if err := db.Preload("Items").Where("id = ?", id).Take(&txn).Error; err != nil {
		return nil, err
	}
	return &txn, nil
}

// FindByStoreId returns paginated transactions for a specific store.
func (r *TransactionRepository) FindByStoreId(db *gorm.DB, storeId string, page int, size int) ([]entity.Transaction, int64, error) {
	var transactions []entity.Transaction
	var total int64

	if err := db.Model(&entity.Transaction{}).Where("store_id = ?", storeId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Items").Where("store_id = ?", storeId).
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// ListByStoreAndDate mengambil daftar transaksi milik sebuah toko pada tanggal tertentu
func (r *TransactionRepository) ListByStoreAndDate(db *gorm.DB, storeId string, date string) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := db.Where("store_id = ? AND transaction_date = ?", storeId, date).Find(&transactions).Error
	return transactions, err
}