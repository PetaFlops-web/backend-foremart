package transaction

import (
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/repository"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/mlclient"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Module struct {
	Controller *controller.TransactionController
	client     *clientImpl
	db         *gorm.DB
}

func New(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	config *viper.Viper,
	productClient product_client.Client,
) *Module {
	mlBaseURL := config.GetString("ML_SERVICE_URL")
	if mlBaseURL == "" {
		mlBaseURL = "http://127.0.0.1:8000" // Fallback default
	}

	mlClient := mlclient.NewMLClient(mlBaseURL)

	transactionRepo := repository.NewTransactionRepository(log)
	transactionItemRepo := repository.NewTransactionItemRepository(log)

	transactionUseCase := usecase.NewTransactionUseCase(
		db, log, validate, transactionRepo, transactionItemRepo, productClient, mlClient,
	)

	transactionController := controller.NewTransactionController(transactionUseCase, log)

	return &Module{
		Controller: transactionController,
		client: &clientImpl{
			db:                  db,
			transactionRepo:     transactionRepo,
			transactionItemRepo: transactionItemRepo,
		},
		db: db,
	}
}

func (m *Module) Client() transaction_client.Client {
	return m.client
}

func (m *Module) Migrate() error {
	if err := m.db.AutoMigrate(&entity.Transaction{}); err != nil {
		return err
	}
	return m.db.AutoMigrate(&entity.TransactionItem{})
}