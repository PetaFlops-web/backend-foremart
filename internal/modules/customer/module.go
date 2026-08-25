package customer

import (
	customer_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer-client"
	"github.com/gofiber/fiber/v2"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/repository"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Module struct {
	Controller *controller.CustomerController
	client     *clientImpl
	db         *gorm.DB
}

func New(db *gorm.DB, log *logrus.Logger, validate *validator.Validate) *Module {
	customerRepo := repository.NewCustomerRepository(log)
	customerUseCase := usecase.NewCustomerUseCase(db, log, validate, customerRepo)
	customerController := controller.NewCustomerController(customerUseCase, log)

	return &Module{
		Controller: customerController,
		client:     &clientImpl{db: db, customerRepo: customerRepo},
		db:         db,
	}
}

func (m *Module) Client() customer_client.Client {
	return m.client
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(&entity.Customer{})
}

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	RegisterRoutes(router, authMiddleware)
}
