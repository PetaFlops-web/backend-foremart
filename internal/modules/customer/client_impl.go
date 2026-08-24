package customer

import (
	"context"

	customer_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/repository"
	"gorm.io/gorm"
)

type clientImpl struct {
	db           *gorm.DB
	customerRepo *repository.CustomerRepository
}

func (c *clientImpl) GetByID(ctx context.Context, id int) (*customer_client.CustomerDTO, error) {
	cust := new(entity.Customer)
	if err := c.customerRepo.FindById(c.db.WithContext(ctx), cust, id); err != nil {
		return nil, err
	}
	return mapToDTO(cust), nil
}

func (c *clientImpl) FindByPhone(ctx context.Context, phone string) (*customer_client.CustomerDTO, error) {
	cust, err := c.customerRepo.FindByPhone(c.db.WithContext(ctx), phone)
	if err != nil {
		return nil, err
	}
	return mapToDTO(cust), nil
}

func (c *clientImpl) ListByStoreID(ctx context.Context, storeID string) ([]customer_client.CustomerDTO, error) {
	customers, err := c.customerRepo.ListByStoreID(c.db.WithContext(ctx), storeID)
	if err != nil {
		return nil, err
	}

	dtos := make([]customer_client.CustomerDTO, len(customers))
	for i, cust := range customers {
		dtos[i] = *mapToDTO(&cust)
	}
	return dtos, nil
}

func mapToDTO(c *entity.Customer) *customer_client.CustomerDTO {
	return &customer_client.CustomerDTO{
		ID:               c.ID,
		Name:             c.Name,
		Phone:            c.Phone,
		CreatedByStoreID: c.CreatedByStoreID,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}
