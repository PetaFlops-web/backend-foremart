package converter

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/model"
)

// TransactionToResponse converts a Transaction entity (with items) to a TransactionResponse.
func TransactionToResponse(txn *entity.Transaction) *model.TransactionResponse {
	itemResponses := make([]model.TransactionItemResponse, len(txn.Items))
	for i, item := range txn.Items {
		itemResponses[i] = *TransactionItemToResponse(&item)
	}

	return &model.TransactionResponse{
		ID:              txn.ID,
		StoreID:         txn.StoreID,
		TransactionDate: txn.TransactionDate.Format("2006-01-02"),
		Source:          txn.Source,
		CreatedAt:       txn.CreatedAt,
		Items:           itemResponses,
	}
}