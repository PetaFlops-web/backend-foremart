package usecase

import (
	"context"
	"time"

	customer_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer-client"
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/model"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/mlclient"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SurvivalUseCase struct {
	Log               *logrus.Logger
	Validate          *validator.Validate
	ProductClient     product_client.Client
	CustomerClient    customer_client.Client
	TransactionClient transaction_client.Client
	MLClient          mlclient.MLClient
}

func NewSurvivalUseCase(
	log *logrus.Logger,
	validate *validator.Validate,
	productClient product_client.Client,
	customerClient customer_client.Client,
	transactionClient transaction_client.Client,
	mlClient mlclient.MLClient,
) *SurvivalUseCase {
	return &SurvivalUseCase{
		Log:               log,
		Validate:          validate,
		ProductClient:     productClient,
		CustomerClient:    customerClient,
		TransactionClient: transactionClient,
		MLClient:          mlClient,
	}
}

func (u *SurvivalUseCase) Predict(ctx context.Context, req *model.SurvivalPredictionRequest) (*model.SurvivalPredictionResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.Warnf("Invalid survival prediction request : %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Data request tidak valid")
	}

	product, err := u.ProductClient.GetByID(ctx, req.ProductID)
	if err != nil {
		u.Log.Warnf("Product %s not found: %+v", req.ProductID, err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Produk tidak ditemukan")
	}
	if product.StoreID != req.StoreID {
		return nil, fiber.NewError(fiber.StatusForbidden, "Produk tidak termasuk dalam toko Anda")
	}

	customer, err := u.CustomerClient.GetByID(ctx, req.CustomerID)
	if err != nil {
		u.Log.Warnf("Customer %s not found: %+v", req.CustomerID, err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Customer tidak ditemukan")
	}

	purchases, err := u.TransactionClient.ListPurchasesByCustomerProduct(ctx, req.StoreID, req.CustomerID, req.ProductID)
	if err != nil {
		u.Log.Warnf("Failed to load purchase history : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil riwayat pembelian")
	}
	if len(purchases) == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "Tidak ada riwayat pembelian produk ini")
	}

	last := purchases[len(purchases)-1]

	purchaseNumber := len(purchases)
	quantity := float64(last.Qty)
	unitPrice := float64(last.SellingPriceSnapshot)
	month := int(last.TransactionDate.Month())
	dayOfWeek := (int(last.TransactionDate.Weekday()) + 6) % 7

	daysSincePrev := 0.0
	avgDaysBetween := 0.0
	if len(purchases) >= 2 {
		gaps := make([]float64, 0, len(purchases)-1)
		for i := 0; i < len(purchases)-1; i++ {
			gaps = append(gaps, float64(calendarDayDiff(purchases[i].TransactionDate, purchases[i+1].TransactionDate)))
		}
		daysSincePrev = gaps[len(gaps)-1]
		total := 0.0
		for _, g := range gaps {
			total += g
		}
		avgDaysBetween = total / float64(len(gaps))
	}

	daysSinceLastBuy := calendarDayDiff(last.TransactionDate, time.Now())
	if daysSinceLastBuy < 0 {
		daysSinceLastBuy = 0
	}
	lastInvoiceDate := last.TransactionDate.Format("2006-01-02")

	items, err := u.TransactionClient.ListItemsByTransaction(ctx, last.TransactionID)
	if err != nil {
		u.Log.Warnf("Failed to load transaction items : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil detail transaksi")
	}

	basketSize := 0.0
	basketValue := 0.0
	uniqueProductIDs := make(map[string]struct{})
	for _, item := range items {
		basketSize += float64(item.Qty)
		basketValue += float64(item.Qty) * float64(item.SellingPriceSnapshot)
		uniqueProductIDs[item.ProductID] = struct{}{}
	}
	basketUnique := len(uniqueProductIDs)

	// ML service expects an integer customer_id; use the numeric surrogate key.
	mlReq := mlclient.SurvivalPredictionRequest{
		CustomerID:       customer.NumericID,
		StockCode:        product.ProductName,
		Quantity:         quantity,
		UnitPrice:        unitPrice,
		BasketSize:       basketSize,
		BasketUnique:     basketUnique,
		BasketValue:      basketValue,
		PurchaseNumber:   purchaseNumber,
		DaysSincePrev:    daysSincePrev,
		AvgDaysBetween:   avgDaysBetween,
		Month:            month,
		DayOfWeek:        dayOfWeek,
		DaysSinceLastBuy: daysSinceLastBuy,
		LastInvoiceDate:  lastInvoiceDate,
	}

	mlResp, err := u.MLClient.PredictSurvival(ctx, mlReq)
	if err != nil {
		u.Log.Errorf("Failed calling ML service: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal memanggil layanan prediksi")
	}

	return &model.SurvivalPredictionResponse{
		CustomerID:             req.CustomerID,
		StockCode:              mlResp.StockCode,
		PurchaseNumber:         purchaseNumber,
		PredictedRestockDate:   mlResp.PredictedRestockDate,
		PredDaysLeft:           mlResp.PredDaysLeft,
		PredMedianSurvivalDays: mlResp.PredMedianSurvivalDays,
		DaysSinceLastBuy:       mlResp.DaysSinceLastBuy,
		ProbBuyWithin7d:        mlResp.ProbBuyWithin7d,
		ProbBuyWithin14d:       mlResp.ProbBuyWithin14d,
		ProbBuyWithin30d:       mlResp.ProbBuyWithin30d,
		PartialHazard:          mlResp.PartialHazard,
	}, nil
}

func calendarDayDiff(from, to time.Time) int {
	f := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	t := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(t.Sub(f).Hours() / 24)
}
