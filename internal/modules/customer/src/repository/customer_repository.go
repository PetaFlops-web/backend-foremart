package repository

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	repository.Repository[entity.Customer]
	Log *logrus.Logger
}

func NewCustomerRepository(log *logrus.Logger) *CustomerRepository {
	return &CustomerRepository{Log: log}
}

func (r *CustomerRepository) FindByPhone(db *gorm.DB, phone string) (*entity.Customer, error) {
	var customer entity.Customer
	if err := db.Where("phone = ?", phone).Take(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) SearchScoped(db *gorm.DB, request *model.SearchCustomerRequest) ([]entity.Customer, int64, error) {
	storeScope := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("(created_by_store_id = ? OR id IN (SELECT DISTINCT customer_id FROM transactions WHERE store_id = ? AND customer_id IS NOT NULL))", request.StoreID, request.StoreID)
	}

	var customers []entity.Customer
	if err := db.Scopes(storeScope, r.FilterCustomer(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Customer{}).Scopes(storeScope, r.FilterCustomer(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *CustomerRepository) ListByStoreID(db *gorm.DB, storeID string) ([]entity.Customer, error) {
	var customers []entity.Customer
	storeScope := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("(created_by_store_id = ? OR id IN (SELECT DISTINCT customer_id FROM transactions WHERE store_id = ? AND customer_id IS NOT NULL))", storeID, storeID)
	}
	if err := db.Scopes(storeScope).Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *CustomerRepository) FilterCustomer(request *model.SearchCustomerRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if request.Search != "" {
			search := "%" + request.Search + "%"
			tx = tx.Where("(name LIKE ? OR phone LIKE ?)", search, search)
		}
		return tx
	}
}
