package transaction

import (
	"context"

	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/repository"
	"gorm.io/gorm"
)

type clientImpl struct {
	db                  *gorm.DB
	transactionRepo     *repository.TransactionRepository
	transactionItemRepo *repository.TransactionItemRepository
}

func (c *clientImpl) ListByStoreAndDate(ctx context.Context, storeId string, date string) ([]transaction_client.TransactionDTO, error) {
	txns, err := c.transactionRepo.ListByStoreAndDate(c.db.WithContext(ctx), storeId, date)
	if err != nil {
		return nil, err
	}

	dtos := make([]transaction_client.TransactionDTO, len(txns))
	for i, t := range txns {
		dtos[i] = transaction_client.TransactionDTO{
			ID:              t.ID,
			StoreID:         t.StoreID,
			TransactionDate: t.TransactionDate,
			Source:          t.Source,
			CreatedAt:       t.CreatedAt,
		}
	}
	return dtos, nil
}

func (c *clientImpl) ListItemsByStoreAndDate(ctx context.Context, storeId string, date string) ([]transaction_client.TransactionItemDTO, error) {
	items, err := c.transactionItemRepo.ListItemsByStoreAndDate(c.db.WithContext(ctx), storeId, date)
	if err != nil {
		return nil, err
	}

	return mapItemsToDTO(items), nil
}

func (c *clientImpl) ListItemsByProduct(ctx context.Context, productId string, lookbackDays int) ([]transaction_client.TransactionItemDTO, error) {
	items, err := c.transactionItemRepo.ListItemsByProduct(c.db.WithContext(ctx), productId, lookbackDays)
	if err != nil {
		return nil, err
	}

	return mapItemsToDTO(items), nil
}

func (c *clientImpl) SumQtyByProductInMonth(ctx context.Context, productId string, yearMonth string) (int, error) {
	return c.transactionItemRepo.SumQtyByProductInMonth(c.db.WithContext(ctx), productId, yearMonth)
}

func mapItemsToDTO(items []entity.TransactionItem) []transaction_client.TransactionItemDTO {
	dtos := make([]transaction_client.TransactionItemDTO, len(items))
	for i, item := range items {
		dtos[i] = transaction_client.TransactionItemDTO{
			ID:                   item.ID,
			TransactionID:        item.TransactionID,
			ProductID:            item.ProductID,
			ProductNameSnapshot:  item.ProductNameSnapshot,
			Qty:                  item.Qty,
			CostPriceSnapshot:    item.CostPriceSnapshot,
			SellingPriceSnapshot: item.SellingPriceSnapshot,
		}
	}
	return dtos
}