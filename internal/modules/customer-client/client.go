package customer_client

import "context"

type CustomerDTO struct {
	ID               string
	Name             string
	Phone            string
	CreatedByStoreID string
	CreatedAt        int64
	UpdatedAt        int64
}

type Client interface {
	GetByID(ctx context.Context, id string) (*CustomerDTO, error)
	FindByPhone(ctx context.Context, phone string) (*CustomerDTO, error)
}
