package usecase

import (
	"context"
	"errors"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/model/converter"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerUseCase struct {
	DB                 *gorm.DB
	Log                *logrus.Logger
	Validate           *validator.Validate
	CustomerRepository *repository.CustomerRepository
}

func NewCustomerUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	customerRepo *repository.CustomerRepository,
) *CustomerUseCase {
	return &CustomerUseCase{
		DB:                 db,
		Log:                log,
		Validate:           validate,
		CustomerRepository: customerRepo,
	}
}

func (u *CustomerUseCase) Create(ctx context.Context, request *model.CreateCustomerRequest) (*model.CustomerResponse, error) {
	if err := u.Validate.Struct(request); err != nil {
		u.Log.Warnf("Invalid request body : %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Data customer tidak valid")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Upsert by phone: return existing customer as-is if phone already exists.
	existing, err := u.CustomerRepository.FindByPhone(tx, request.Phone)
	if err == nil && existing != nil {
		if err := tx.Commit().Error; err != nil {
			u.Log.Warnf("Failed commit transaction : %+v", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan customer")
		}
		return converter.CustomerToResponse(existing), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		u.Log.Warnf("Failed find customer by phone : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan customer")
	}

	customer := &entity.Customer{
		ID:               uuid.New().String(),
		Name:             request.Name,
		Phone:            request.Phone,
		CreatedByStoreID: request.StoreID,
	}

	if err := u.CustomerRepository.Create(tx, customer); err != nil {
		u.Log.Warnf("Failed create customer : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan customer")
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan customer")
	}

	return converter.CustomerToResponse(customer), nil
}

func (u *CustomerUseCase) Get(ctx context.Context, storeId, customerID string) (*model.CustomerResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	customer := new(entity.Customer)
	if err := u.CustomerRepository.FindById(tx, customer, customerID); err != nil {
		u.Log.Warnf("Customer not found : %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Customer tidak ditemukan")
	}

	if customer.CreatedByStoreID != storeId {
		u.Log.Warnf("Customer not found: store mismatch")
		return nil, fiber.NewError(fiber.StatusNotFound, "Customer tidak ditemukan")
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data customer")
	}

	return converter.CustomerToResponse(customer), nil
}

func (u *CustomerUseCase) Search(ctx context.Context, request *model.SearchCustomerRequest) ([]model.CustomerResponse, int64, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(request); err != nil {
		u.Log.Warnf("Invalid request query : %+v", err)
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, "Format pencarian tidak valid")
	}

	if request.StoreID == "" {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, "Store ID wajib diisi")
	}

	customers, total, err := u.CustomerRepository.SearchScoped(tx, request)
	if err != nil {
		u.Log.Warnf("Failed search customers : %+v", err)
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, "Gagal mencari customer")
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, "Gagal mencari customer")
	}

	responses := make([]model.CustomerResponse, len(customers))
	for i, customer := range customers {
		responses[i] = *converter.CustomerToResponse(&customer)
	}

	return responses, total, nil
}
