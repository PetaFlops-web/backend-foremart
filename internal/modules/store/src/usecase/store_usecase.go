package usecase

import (
	"context"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/model/converter"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StoreUseCase struct {
	DB 			 *gorm.DB
	Log 		 *logrus.Logger
	Validate 	 *validator.Validate
	StoreRepo    *repository.StoreRepository
}

func NewStoreUseCase(
	db *gorm.DB, log *logrus.Logger, validate *validator.Validate, storeRepo *repository.StoreRepository) *StoreUseCase {
	return &StoreUseCase{
		DB:        db,
		Log:       log,
		Validate:  validate,
		StoreRepo: storeRepo,
	}
}

func (u *StoreUseCase) Create(ctx context.Context, request *model.StoreRequest) (*model.StoreResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(request); err != nil {
		u.Log.Warnf("Invalid request body : %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	// Check if user already has a store
	existingStore := new(entity.Store)
	err := u.StoreRepo.FindByUserID(tx, existingStore, request.UserID)
	if err == nil {
		// User already has a store - return it instead of error
		u.Log.Infof("User %s already has a store: %s", request.UserID, existingStore.ID)
		tx.Rollback()
		return converter.StoreToResponse(existingStore), nil
	}

	store := &entity.Store{
		ID:        uuid.New().String(),
		UserID:    request.UserID,
		StoreName: request.StoreName,
	}

	if err := u.StoreRepo.Create(tx, store); err != nil {
		u.Log.Warnf("Failed to create store : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat toko")
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan data toko")
	}

	return converter.StoreToResponse(store), nil
}

func (u *StoreUseCase) Get(ctx context.Context, userID string, storeID string) (*model.StoreResponse, error) {
	tx := u.DB.WithContext(ctx)

	store := new(entity.Store)
	if err := u.StoreRepo.FindByIdAndUserID(tx, store, storeID, userID); err != nil {
		u.Log.Warnf("Failed to find store : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mencari data toko")
	}

	return converter.StoreToResponse(store), nil
}

func (u *StoreUseCase) GetByUserID(ctx context.Context, userID string) (*model.StoreResponse, error) {
	tx := u.DB.WithContext(ctx)
	store := new(entity.Store)
	if err := u.StoreRepo.FindByUserID(tx, store, userID); err != nil {
		u.Log.Warnf("Failed to find store by user ID: %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "User belum memiliki toko")
	}
	return converter.StoreToResponse(store), nil
}

func (u *StoreUseCase) Update(ctx context.Context, request *model.StoreRequest) (*model.StoreResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(request); err != nil {
		u.Log.Warnf("Invalid request body : %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	store := new(entity.Store)
	if err := u.StoreRepo.FindByIdAndUserID(tx, store, request.ID, request.UserID); err != nil {
		u.Log.Warnf("Failed to find store : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mencari data toko")
	}

	store.StoreName = request.StoreName

	if err := u.StoreRepo.Update(tx, store); err != nil {
		u.Log.Warnf("Failed to update store : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal memperbarui data toko")
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan data toko")
	}

	return converter.StoreToResponse(store), nil
}

