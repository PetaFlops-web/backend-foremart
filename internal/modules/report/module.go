package report

import (
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/report/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/report/src/usecase"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type Module struct {
	Controller *controller.ReportController
}

func New(
	log *logrus.Logger,
	validate *validator.Validate,
	transactionClient transaction_client.Client,
	productClient product_client.Client,
	storeClient store_client.StoreClient,
) *Module {
	reportUseCase := usecase.NewReportUseCase(log, validate, transactionClient, productClient, storeClient)
	reportController := controller.NewReportController(reportUseCase, log)

	return &Module{
		Controller: reportController,
	}
}

func (m *Module) Migrate() error {
	return nil
}
