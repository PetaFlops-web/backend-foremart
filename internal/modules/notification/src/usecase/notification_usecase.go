package usecase

import (
	"context"
	"fmt"
	"time"

	customer_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/model/converter"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/repository"
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	survival_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival-client"
	survival_model "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/model"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/notifier"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// DaysThreshold is how many days before the predicted reorder date a
	// reminder is sent.
	DaysThreshold = 3

	// Promo thresholds: number of purchases required for each discount tier.
	PromoTier3Threshold = 3
	PromoTier5Threshold = 5
)

// ruleForPromo maps a purchase number to a discount percentage (0 = no promo).
func ruleForPromo(purchaseNumber int) int {
	if purchaseNumber >= PromoTier5Threshold {
		return 30
	}
	if purchaseNumber >= PromoTier3Threshold {
		return 20
	}
	return 0
}

type NotificationUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Repo              *repository.NotificationRepository
	StoreClient       store_client.StoreClient
	CustomerClient    customer_client.Client
	ProductClient     product_client.Client
	TransactionClient transaction_client.Client
	SurvivalClient    survival_client.Client
	Notifier          notifier.Notifier
}

func NewNotificationUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	repo *repository.NotificationRepository,
	storeClient store_client.StoreClient,
	customerClient customer_client.Client,
	productClient product_client.Client,
	transactionClient transaction_client.Client,
	survivalClient survival_client.Client,
	notifier notifier.Notifier,
) *NotificationUseCase {
	return &NotificationUseCase{
		DB:                db,
		Log:               log,
		Repo:              repo,
		StoreClient:       storeClient,
		CustomerClient:    customerClient,
		ProductClient:     productClient,
		TransactionClient: transactionClient,
		SurvivalClient:    survivalClient,
		Notifier:          notifier,
	}
}

// RunReminder iterates every store's customers and products, predicts reorder
// timing via the survival service, and sends a WhatsApp reminder with an
// optional promo for customers due to repurchase soon.
func (u *NotificationUseCase) RunReminder(ctx context.Context) (int, error) {
	stores, err := u.StoreClient.ListStores(ctx)
	if err != nil {
		return 0, fmt.Errorf("list stores: %w", err)
	}

	period := time.Now().Format("2006-01-02")
	sent := 0

	for _, store := range stores {
		customers, err := u.CustomerClient.ListByStoreID(ctx, store.ID)
		if err != nil {
			u.Log.WithError(err).WithField("store_id", store.ID).Warn("list customers failed; skipping store")
			continue
		}
		products, err := u.ProductClient.ListByStoreID(ctx, store.ID)
		if err != nil {
			u.Log.WithError(err).WithField("store_id", store.ID).Warn("list products failed; skipping store")
			continue
		}

		for _, cust := range customers {
			for _, prod := range products {
				if u.processPair(ctx, store.ID, cust, prod, period) {
					sent++
				}
			}
		}
	}

	return sent, nil
}

func (u *NotificationUseCase) processPair(ctx context.Context, storeID string, cust customer_client.CustomerDTO, prod product_client.ProductDTO, period string) bool {
	// Skip if already notified this period (dedup).
	exists, err := u.Repo.ExistsForPeriod(u.DB.WithContext(ctx), cust.ID, prod.ID, period)
	if err != nil {
		u.Log.WithError(err).Warn("dedup check failed")
		return false
	}
	if exists {
		return false
	}

	// Skip customers with no purchase history for this product.
	purchases, err := u.TransactionClient.ListPurchasesByCustomerProduct(ctx, storeID, cust.ID, prod.ID)
	if err != nil {
		u.Log.WithError(err).Warn("load purchase history failed")
		return false
	}
	if len(purchases) == 0 {
		return false
	}

	pred, err := u.SurvivalClient.Predict(ctx, &survival_model.SurvivalPredictionRequest{
		StoreID:    storeID,
		CustomerID: cust.ID,
		ProductID:  prod.ID,
	})
	if err != nil {
		u.Log.WithError(err).WithFields(logrus.Fields{
			"customer_id": cust.ID,
			"product_id":  prod.ID,
		}).Warn("survival prediction failed")
		return false
	}

	if pred.PredDaysLeft > DaysThreshold {
		return false
	}

	discount := ruleForPromo(pred.PurchaseNumber)
	rule := "REMINDER"
	message := buildMessage(cust.Name, prod.ProductName, pred.PredictedRestockDate, discount)
	if discount > 0 {
		rule = "REPEAT_3X"
	}

	// Send first, then log the result so failed sends are recorded too.
	status := "sent"
	if err := u.Notifier.SendReminder(ctx, notifier.Reminder{To: cust.Phone, Message: message}); err != nil {
		u.Log.WithError(err).Warn("send reminder failed")
		status = "failed"
	}

	id, err := utils.GenerateNotificationId()
	if err != nil {
		u.Log.WithError(err).Warn("generate notification id failed")
		return false
	}

	logEntry := &entity.NotificationLog{
		ID:                   id,
		StoreID:              storeID,
		CustomerID:           cust.ID,
		ProductID:            prod.ID,
		Channel:              "whatsapp",
		Message:              message,
		PredictedRestockDate: pred.PredictedRestockDate,
		RuleTriggered:        rule,
		Status:               status,
		Period:               period,
	}
	if err := u.Repo.Create(u.DB.WithContext(ctx), logEntry); err != nil {
		u.Log.WithError(err).Warn("insert notification log failed")
		return false
	}

	return true
}

func buildMessage(name string, productName string, predictedDate string, discount int) string {
	greeting := "Halo"
	if name != "" {
		greeting = fmt.Sprintf("Halo %s", name)
	}

	if discount > 0 {
		return fmt.Sprintf(
			"%s, persediaan %s Anda mungkin sudah mulai habis. Anda dapat membeli kembali %s dengan potongan harga %d%% karena Anda sudah berbelanja beberapa kali. Prediksi waktu pembelian ulang Anda: %s. Sampai jumpa!",
			greeting, productName, productName, discount, predictedDate,
		)
	}

	return fmt.Sprintf(
		"%s, persediaan %s Anda mungkin sudah mulai habis. Jangan lupa untuk membeli kembali %s ya. Prediksi waktu pembelian ulang Anda: %s.",
		greeting, productName, productName, predictedDate,
	)
}

// List returns stored notification logs for a store.
func (u *NotificationUseCase) List(ctx context.Context, storeID string) ([]model.NotificationResponse, error) {
	logs, err := u.Repo.ListByStoreID(u.DB.WithContext(ctx), storeID)
	if err != nil {
		return nil, err
	}

	resp := make([]model.NotificationResponse, len(logs))
	for i := range logs {
		resp[i] = *converter.NotificationToResponse(&logs[i])
	}
	return resp, nil
}
