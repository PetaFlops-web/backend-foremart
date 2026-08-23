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

// CustomerProductPurchase is a single purchase of a specific product by a
// specific customer, joined with its transaction date. Used by the Survival
// module to build ML features.
type CustomerProductPurchase struct {
	TransactionID        string    `gorm:"column:transaction_id"`
	TransactionDate      time.Time `gorm:"column:transaction_date"`
	Qty                  int       `gorm:"column:qty"`
	SellingPriceSnapshot int64     `gorm:"column:selling_price_snapshot"`
}

// ListPurchasesByCustomerProduct returns every purchase of a product by a
// customer in a store, ordered oldest → newest.
func (r *TransactionItemRepository) ListPurchasesByCustomerProduct(db *gorm.DB, storeID string, customerID int, productID string) ([]CustomerProductPurchase, error) {
	var rows []CustomerProductPurchase
	err := db.Table("transaction_items").
		Select("transaction_items.transaction_id, transactions.transaction_date, transaction_items.qty, transaction_items.selling_price_snapshot").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("transactions.store_id = ? AND transactions.customer_id = ? AND transaction_items.product_id = ?", storeID, customerID, productID).
		Order("transactions.transaction_date ASC, transaction_items.transaction_id ASC").
		Find(&rows).Error
	return rows, err
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

// DailySalesByProduct is a row in the zero-filled daily-sales series.
type DailySalesByProduct struct {
	SaleDate string  `gorm:"column:sale_date"`
	TotalQty float64 `gorm:"column:total_qty"`
}

// GetDailySalesHistoryByProduct returns a complete daily sales series for a
// product in a store over the window [endDate - (historyDays-1) .. endDate],
// ordered oldest → newest. Days with no sales are filled with 0 so the series
// always has exactly historyDays entries (one per calendar day).
func (r *TransactionItemRepository) GetDailySalesHistoryByProduct(db *gorm.DB, storeID string, productID string, historyDays int, endDate string) ([]DailySalesByProduct, error) {
	parsedEnd, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	if historyDays < 1 {
		return nil, nil
	}

	start := parsedEnd.AddDate(0, 0, -(historyDays - 1))

	var rows []DailySalesByProduct
	err = db.Model(&entity.TransactionItem{}).
		Select("DATE_FORMAT(transactions.transaction_date, '%Y-%m-%d') AS sale_date, COALESCE(SUM(transaction_items.qty), 0) AS total_qty").
		Joins("JOIN transactions ON transaction_items.transaction_id = transactions.id").
		Where("transactions.store_id = ? AND transaction_items.product_id = ? AND transactions.transaction_date >= ? AND transactions.transaction_date <= ?",
			storeID, productID, start.Format("2006-01-02"), endDate).
		Group("transactions.transaction_date").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	salesByDate := make(map[string]float64, len(rows))
	for _, row := range rows {
		salesByDate[row.SaleDate] = row.TotalQty
	}

	series := make([]DailySalesByProduct, 0, historyDays)
	for i := 0; i < historyDays; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		series = append(series, DailySalesByProduct{
			SaleDate: day,
			TotalQty: salesByDate[day],
		})
	}

	return series, nil
}