package store

import (
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/repository"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Module struct {
	Controller *controller.StoreController
	client     *clientImpl
	db         *gorm.DB
}

func New(db *gorm.DB, validate *validator.Validate, log *logrus.Logger) *Module {
	storeRepo := repository.NewStoreRepository(log)
	storeUseCase := usecase.NewStoreUseCase(db, log, validate, storeRepo)
	storeController := controller.NewStoreController(storeUseCase, log)

	return &Module{
		Controller: storeController,
		client:     &clientImpl{repo: *storeRepo, db: db},
		db:         db,
	}
}

func (m *Module) Client() store_client.StoreClient {
	return m.client
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(&entity.Store{})
}
