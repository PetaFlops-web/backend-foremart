package store_client

import "context"

type StoreDTO struct {
	ID        string
	UserID    string
	StoreName string
	CreatedAt int64
	UpdatedAt int64
}

type StoreClient interface {
	GetStoreByID(ctx context.Context, storeID string) (*StoreDTO, error)
	GetStoreByUserID(ctx context.Context, userID string) (*StoreDTO, error)
}
