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

type Client interface {
	ListByStoreAndDate(ctx context.Context, storeId string, date string) ([]TransactionDTO, error)
	ListItemsByStoreAndDate(ctx context.Context, storeId string, date string) ([]TransactionItemDTO, error)
	ListItemsByProduct(ctx context.Context, productId string, lookbackDays int) ([]TransactionItemDTO, error)
	SumQtyByProductInMonth(ctx context.Context, productId string, yearMonth string) (int, error)
}