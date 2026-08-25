package customer_client

import "context"

type CustomerDTO struct {
	ID               string
	NumericID        int64
	Name             string
	Phone            string
	CreatedByStoreID string
	CreatedAt        int64
	UpdatedAt        int64
}

type Client interface {
	GetByID(ctx context.Context, id string) (*CustomerDTO, error)
	FindByPhone(ctx context.Context, phone string) (*CustomerDTO, error)
	ListByStoreID(ctx context.Context, storeID string) ([]CustomerDTO, error)
}
