package converter

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/model"
)

func CustomerToResponse(c *entity.Customer) *model.CustomerResponse {
	if c == nil {
		return nil
	}
	return &model.CustomerResponse{
		ID:               c.ID,
		NumericID:        c.NumericID,
		Name:             c.Name,
		Phone:            c.Phone,
		CreatedByStoreID: c.CreatedByStoreID,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}
