package transaction_client

import (
	"context"
	"time"
)

type TransactionDTO struct {
	ID              string
	StoreID         string
	TransactionDate time.Time
	Source          string
	CreatedAt       int64
}

type TransactionItemDTO struct {
	ID                   string
	TransactionID        string
	ProductID            string
	ProductNameSnapshot  string
	Qty                  int
	CostPriceSnapshot    int64
	SellingPriceSnapshot int64
}

type CustomerProductPurchaseDTO struct {
	TransactionID        string
	TransactionDate      time.Time
	Qty                  int
	SellingPriceSnapshot int64
}

type Client interface {
	ListByStoreAndDate(ctx context.Context, storeId string, date string) ([]TransactionDTO, error)
	ListItemsByStoreAndDate(ctx context.Context, storeId string, date string) ([]TransactionItemDTO, error)
	ListItemsByProduct(ctx context.Context, productId string, lookbackDays int) ([]TransactionItemDTO, error)
	SumQtyByProductInMonth(ctx context.Context, productId string, yearMonth string) (int, error)
	GetDailySalesHistoryByProduct(ctx context.Context, storeID string, productID string, historyDays int, endDate string) ([]float64, error)
	ListPurchasesByCustomerProduct(ctx context.Context, storeID string, customerID int, productID string) ([]CustomerProductPurchaseDTO, error)
	ListItemsByTransaction(ctx context.Context, transactionID string) ([]TransactionItemDTO, error)
}