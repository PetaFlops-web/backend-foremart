package restock

import (
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/repository"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/usecase"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/mlclient"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Module struct {
	Controller *controller.RestockController
	db         *gorm.DB
}

func New(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	config *viper.Viper,
	storeClient store_client.StoreClient,
	productClient product_client.Client,
	transactionClient transaction_client.Client,
) *Module {
	mlBaseURL := config.GetString("ML_SERVICE_URL")
	if mlBaseURL == "" {
		mlBaseURL = "http://127.0.0.1:8000"
	}

	restockRepo := repository.NewRestockRepository(log, db)
	restockUseCase := usecase.NewRestockUseCase(
		log,
		validate,
		restockRepo,
		storeClient,
		productClient,
		transactionClient,
		mlclient.NewMLClient(mlBaseURL),
	)
	restockController := controller.NewRestockController(restockUseCase, log)

	return &Module{
		Controller: restockController,
		db:         db,
	}
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(&entity.RestockPrediction{})
}
