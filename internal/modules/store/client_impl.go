package store

import (
	"context"

	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/repository"
	"gorm.io/gorm"
)

type clientImpl struct {
	repo repository.StoreRepository
	db   *gorm.DB
}

func (c *clientImpl) GetStoreByID(ctx context.Context, storeID string) (*store_client.StoreDTO, error) {
	store := new(entity.Store)
	if err := c.repo.FindById(c.db.WithContext(ctx), store, storeID); err != nil {
		return nil, err
	}

	return mapToDTO(store), nil
}

func (c *clientImpl) GetStoreByUserID(ctx context.Context, userID string) (*store_client.StoreDTO, error) {
	store := new(entity.Store)
	if err := c.repo.FindByUserID(c.db.WithContext(ctx), store, userID); err != nil {
		return nil, err
	}

	return mapToDTO(store), nil
}

func (c *clientImpl) ListStores(ctx context.Context) ([]store_client.StoreDTO, error) {
	stores, err := c.repo.ListAll(c.db.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	dtos := make([]store_client.StoreDTO, len(stores))
	for i, s := range stores {
		dtos[i] = *mapToDTO(&s)
	}
	return dtos, nil
}

func mapToDTO(store *entity.Store) *store_client.StoreDTO {
	return &store_client.StoreDTO{
		ID:        store.ID,
		UserID:    store.UserID,
		StoreName: store.StoreName,
		CreatedAt: store.CreatedAt,
		UpdatedAt: store.UpdatedAt,
	}
}
