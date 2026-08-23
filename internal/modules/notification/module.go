package notification

import (
	customer_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/controller"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/repository"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/usecase"
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	survival_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/notifier"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Module struct {
	Controller *controller.NotificationController
	UseCase    *usecase.NotificationUseCase
	db         *gorm.DB
}

func New(
	db *gorm.DB,
	log *logrus.Logger,
	config *viper.Viper,
	storeClient store_client.StoreClient,
	customerClient customer_client.Client,
	productClient product_client.Client,
	transactionClient transaction_client.Client,
	survivalClient survival_client.Client,
) *Module {
	repo := repository.NewNotificationRepository(log)

	token := config.GetString("fonnte.token")
	target := config.GetString("fonnte.target")
	n, err := notifier.NewNotifier(token, target, log)
	if err != nil {
		log.WithError(err).Warn("notifier fell back to log-only")
	}

	notificationUseCase := usecase.NewNotificationUseCase(
		db, log, repo,
		storeClient, customerClient, productClient, transactionClient,
		survivalClient, n,
	)
	notificationController := controller.NewNotificationController(notificationUseCase, log)

	return &Module{
		Controller: notificationController,
		UseCase:    notificationUseCase,
		db:         db,
	}
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(&entity.NotificationLog{})
}
