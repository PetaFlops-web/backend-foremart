package repository

import (
	"time"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionItemRepository struct {
	repository.Repository[entity.TransactionItem]
	Log *logrus.Logger
}

func NewTransactionItemRepository(log *logrus.Logger) *TransactionItemRepository {
	return &TransactionItemRepository{Log: log}
}

// CreateBatch inserts multiple transaction items at once.
func (r *TransactionItemRepository) CreateBatch(db *gorm.DB, items []entity.TransactionItem) error {
	if len(items) == 0 {
		return nil
	}
	return db.Create(&items).Error
}

// FindByTransactionId retrieves all items for a specific transaction.
func (r *TransactionItemRepository) FindByTransactionId(db *gorm.DB, transactionId string) ([]entity.TransactionItem, error) {
	var items []entity.TransactionItem
	err := db.Where("transaction_id = ?", transactionId).Find(&items).Error
	return items, err
}

// DeleteByTransactionId deletes all items belonging to a transaction.
func (r *TransactionItemRepository) DeleteByTransactionId(db *gorm.DB, transactionId string) error {
	return db.Where("transaction_id = ?", transactionId).Delete(&entity.TransactionItem{}).Error
}

// ListItemsByStoreAndDate mengambil semua item yang terjual di suatu toko pada hari tertentu
// Digunakan oleh modul Report
func (r *TransactionItemRepository) ListItemsByStoreAndDate(db *gorm.DB, storeId string, date string) ([]entity.TransactionItem, error) {
	var items []entity.TransactionItem

	err := db.Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("transactions.store_id = ? AND transactions.transaction_date = ?", storeId, date).
		Find(&items).Error

	return items, err
}

// ListItemsByProduct mengambil histori penjualan sebuah produk spesifik dalam jangka waktu N hari ke belakang
// Digunakan oleh modul Restock
func (r *TransactionItemRepository) ListItemsByProduct(db *gorm.DB, productId string, lookbackDays int) ([]entity.TransactionItem, error) {
	var items []entity.TransactionItem

	// Hitung tanggal batas bawah
	startDate := time.Now().AddDate(0, 0, -lookbackDays).Format("2006-01-02")

	err := db.Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("transaction_items.product_id = ? AND transactions.transaction_date >= ?", productId, startDate).
		Find(&items).Error

	return items, err
}

// SumQtyByProductInMonth menghitung total qty produk spesifik yang terjual di bulan tertentu
// Digunakan oleh modul Promotion
func (r *TransactionItemRepository) SumQtyByProductInMonth(db *gorm.DB, productId string, yearMonth string) (int, error) {
	var totalQty int

	err := db.Model(&entity.TransactionItem{}).
		Select("COALESCE(SUM(transaction_items.qty), 0)").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("transaction_items.product_id = ? AND DATE_FORMAT(transactions.transaction_date, '%Y-%m') = ?", productId, yearMonth).
		Scan(&totalQty).Error

	return totalQty, err
}

// DailySalesByProduct is a row returned by the daily-sales aggregate query.
type DailySalesByProduct struct {
	SaleDate string  `gorm:"column:sale_date"`
	TotalQty float64 `gorm:"column:total_qty"`
}

// GetDailySalesHistoryByProduct aggregates daily total qty sold for a product
// in a store over a date range [endDate - historyDays + 1 .. endDate].
// Returns one row per day ordered oldest → newest; days with no sales get 0
// via a COALESCE fill at the caller level (here we return only actual sale days).
func (r *TransactionItemRepository) GetDailySalesHistoryByProduct(db *gorm.DB, storeID string, productID string, historyDays int, endDate string) ([]DailySalesByProduct, error) {
	var results []DailySalesByProduct

	startDate := time.Now()
	if endDate != "" {
		parsed, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			startDate = parsed
		}
	}
	start := startDate.AddDate(0, 0, -(historyDays - 1)).Format("2006-01-02")

	err := db.Model(&entity.TransactionItem{}).
		Select("transactions.transaction_date AS sale_date, COALESCE(SUM(transaction_items.qty), 0) AS total_qty").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("transactions.store_id = ? AND transaction_items.product_id = ? AND transactions.transaction_date >= ? AND transactions.transaction_date <= ?",
			storeID, productID, start, endDate).
		Group("transactions.transaction_date").
		Order("transactions.transaction_date ASC").
		Scan(&results).Error

	return results, err
}