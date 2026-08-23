package survival

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	survival_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/usecase"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/mlclient"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Module struct {
	Controller *controller.SurvivalController
	client     survival_client.Client
}

func New(
	log *logrus.Logger,
	validate *validator.Validate,
	config *viper.Viper,
	productClient product_client.Client,
	transactionClient transaction_client.Client,
) *Module {
	mlBaseURL := config.GetString("ML_SERVICE_URL")
	if mlBaseURL == "" {
		mlBaseURL = "http://127.0.0.1:8000"
	}

	survivalUseCase := usecase.NewSurvivalUseCase(log, validate, productClient, transactionClient, mlclient.NewMLClient(mlBaseURL))
	survivalController := controller.NewSurvivalController(survivalUseCase, log)

	return &Module{
		Controller: survivalController,
		client:     survivalUseCase,
	}
}

func (m *Module) Client() survival_client.Client {
	return m.client
}

func (m *Module) Migrate() error {
	return nil
}
