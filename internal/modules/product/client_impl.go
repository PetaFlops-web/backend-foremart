package product

import (
	"context"

	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product/src/repository"
	"gorm.io/gorm"
)

type clientImpl struct {
	db          *gorm.DB
	productRepo *repository.ProductRepository
}

func (c *clientImpl) GetByID(ctx context.Context, id string) (*product_client.ProductDTO, error) {
	product := new(entity.Product)
	if err := c.productRepo.FindById(c.db.WithContext(ctx), product, id); err != nil {
		return nil, err
	}
	return mapToDTO(product), nil
}

func (c *clientImpl) ListByStoreID(ctx context.Context, storeId string) ([]product_client.ProductDTO, error) {
	products, err := c.productRepo.FindByStoreId(c.db.WithContext(ctx), storeId)
	if err != nil {
		return nil, err
	}

	dtos := make([]product_client.ProductDTO, len(products))
	for i, product := range products {
		dtos[i] = *mapToDTO(&product)
	}
	return dtos, nil
}

func (c *clientImpl) DecrementStock(ctx context.Context, id string, qty int) error {
	return c.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ? AND stock >= ?", id, qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty)).Error
}

func (c *clientImpl) IncrementStock(ctx context.Context, id string, qty int) error {
	return c.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", qty)).Error
}

func (c *clientImpl) Search(ctx context.Context, storeId string, keyword string) ([]product_client.ProductDTO, error) {
	var products []entity.Product
	tx := c.db.WithContext(ctx).Where("store_id = ?", storeId)
	if keyword != "" {
		keyword = "%" + keyword + "%"
		tx = tx.Where("product_name LIKE ?", keyword)
	}

	if err := tx.Find(&products).Error; err != nil {
		return nil, err
	}

	dtos := make([]product_client.ProductDTO, len(products))
	for i, product := range products {
		dtos[i] = *mapToDTO(&product)
	}
	return dtos, nil
}

func mapToDTO(product *entity.Product) *product_client.ProductDTO {
	return &product_client.ProductDTO{
		ID:           product.ID,
		StoreID:      product.StoreID,
		ProductName:  product.ProductName,
		CostPrice:    product.CostPrice,
		SellingPrice: product.SellingPrice,
		Stock:        product.Stock,
		Unit:         product.Unit,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}
}