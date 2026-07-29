package converter

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/model"
)

// TransactionItemToResponse converts a TransactionItem entity to a TransactionItemResponse.
func TransactionItemToResponse(item *entity.TransactionItem) *model.TransactionItemResponse {
	return &model.TransactionItemResponse{
		ID:                   item.ID,
		ProductID:            item.ProductID,
		ProductNameSnapshot:  item.ProductNameSnapshot,
		Qty:                  item.Qty,
		CostPriceSnapshot:    item.CostPriceSnapshot,
		SellingPriceSnapshot: item.SellingPriceSnapshot,
	}
}