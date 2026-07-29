package product_client

import "context"

type ProductDTO struct {
	ID           string
	StoreID      string
	ProductName  string
	CostPrice    int64
	SellingPrice int64
	Stock        int
	Unit         string
	CreatedAt    int64
	UpdatedAt    int64
}

type Client interface {
	GetByID(ctx context.Context, id string) (*ProductDTO, error)
	ListByStoreID(ctx context.Context, storeId string) ([]ProductDTO, error)
	DecrementStock(ctx context.Context, id string, qty int) error
	IncrementStock(ctx context.Context, id string, qty int) error
	Search(ctx context.Context, storeId string, keyword string) ([]ProductDTO, error)
}