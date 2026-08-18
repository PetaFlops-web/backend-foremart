package usecase

import (
	"context"
	"time"

	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/model/converter"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/repository"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/mlclient"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

const (
	defaultHistoryDays = 30

	skipReasonInsufficientHistory = "riwayat_tidak_cukup"
	skipReasonMLError             = "kesalahan_ml"
	skipReasonNoRestockNeeded     = "restock_tidak_diperlukan"
)

type RestockUseCase struct {
	Log               *logrus.Logger
	Validate          *validator.Validate
	Repo              *repository.RestockRepository
	StoreClient       store_client.StoreClient
	ProductClient     product_client.Client
	TransactionClient transaction_client.Client
	MLClient          mlclient.MLClient
}

func NewRestockUseCase(
	log *logrus.Logger,
	validate *validator.Validate,
	repo *repository.RestockRepository,
	storeClient store_client.StoreClient,
	productClient product_client.Client,
	transactionClient transaction_client.Client,
	mlClient mlclient.MLClient,
) *RestockUseCase {
	return &RestockUseCase{
		Log:               log,
		Validate:          validate,
		Repo:              repo,
		StoreClient:       storeClient,
		ProductClient:     productClient,
		TransactionClient: transactionClient,
		MLClient:          mlClient,
	}
}

func (u *RestockUseCase) Generate(ctx context.Context, req *model.GenerateRestockPredictionRequest) (*model.GenerateRestockPredictionResponse, error) {
	forecastDate, err := u.applyDefaultsAndValidate(req)
	if err != nil {
		return nil, err
	}

	if _, err := u.StoreClient.GetStoreByID(ctx, req.StoreID); err != nil {
		u.Log.Warnf("Store not found : %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Toko tidak ditemukan")
	}

	products, err := u.resolveProducts(ctx, req.StoreID, req.ProductID)
	if err != nil {
		return nil, err
	}

	endDate := forecastDate.AddDate(0, 0, -1).Format("2006-01-02")
	daysAhead := calendarDaysFromToday(*forecastDate)
	if daysAhead < 1 {
		daysAhead = 1
	}

	eligible := make([]productWithHistory, 0, len(products))
	skipped := make([]model.RestockSkippedItemResponse, 0)

	for _, product := range products {
		history, err := u.TransactionClient.GetDailySalesHistoryByProduct(ctx, req.StoreID, product.ID, req.HistoryDays, endDate)
		if err != nil {
			u.Log.Errorf("Failed to get sales history for product %s: %+v", product.ID, err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil histori penjualan")
		}
		if sumFloat(history) == 0 {
			skipped = append(skipped, skippedItem(product, skipReasonInsufficientHistory))
			continue
		}
		eligible = append(eligible, productWithHistory{Product: product, History: history})
	}

	predictions := u.predictEligibleItems(ctx, req, forecastDate.Format("2006-01-02"), eligible, &skipped)
	items := make([]model.RestockPredictionResponse, 0, len(predictions))

	for _, prediction := range predictions {
		product := prediction.Product
		predictedSales := maxInt(0, prediction.Response.PredictedSales)
		neededStock := predictedSales * daysAhead
		recommendedQty := maxInt(0, neededStock - product.Stock)
		if recommendedQty == 0 {
			skipped = append(skipped, skippedItem(product, skipReasonNoRestockNeeded))
			continue
		}

		predictionID, err := utils.GenerateRestockPredictionId()
		if err != nil {
			u.Log.Errorf("Failed to generate restock prediction id: %+v", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat ID prediksi restock")
		}

		entityPrediction := &entity.RestockPrediction{
			ID:                    predictionID,
			StoreID:               req.StoreID,
			ProductID:             product.ID,
			ProductName:           product.ProductName,
			Unit:                  product.Unit,
			ForecastDate:          forecastDate,
			PredictedDailySales:   predictedSales,
			CurrentStock:          product.Stock,
			RecommendedRestockQty: recommendedQty,
			HistoryDays:           req.HistoryDays,
		}

		saved, err := u.Repo.UpsertByStoreProductDate(ctx, entityPrediction)
		if err != nil {
			u.Log.Errorf("Failed to save restock prediction: %+v", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan prediksi restock")
		}
		items = append(items, *converter.RestockPredictionToResponse(saved))
	}

	return &model.GenerateRestockPredictionResponse{
		GeneratedCount: len(items),
		SkippedCount:   len(skipped),
		Items:          items,
		Skipped:        skipped,
	}, nil
}

func (u *RestockUseCase) List(ctx context.Context, storeID string) ([]model.RestockPredictionResponse, error) {
	if storeID == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "store_id wajib diisi")
	}

	if _, err := u.StoreClient.GetStoreByID(ctx, storeID); err != nil {
		u.Log.Warnf("Store not found : %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Toko tidak ditemukan")
	}

	predictions, err := u.Repo.ListByStoreID(ctx, storeID)
	if err != nil {
		u.Log.Errorf("Failed to list restock predictions: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menampilkan prediksi restock")
	}

	return converter.RestockPredictionsToResponses(predictions), nil
}

func (u *RestockUseCase) applyDefaultsAndValidate(req *model.GenerateRestockPredictionRequest) (*time.Time, error) {
	if req.ForecastDate == nil {
		tomorrow := time.Now().AddDate(0, 0, 1)
		req.ForecastDate = &tomorrow
	}
	if req.HistoryDays == 0 {
		req.HistoryDays = defaultHistoryDays
	}

	if err := u.Validate.Struct(req); err != nil {
		u.Log.Warnf("Invalid restock request : %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	if req.HistoryDays < defaultHistoryDays {
		return nil, fiber.NewError(fiber.StatusBadRequest, "history_days minimal 30 hari")
	}

	if calendarDaysFromToday(*req.ForecastDate) < 1 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "forecast_date harus tanggal besok atau setelahnya")
	}

	return req.ForecastDate, nil
}

// calendarDaysFromToday returns the number of calendar days from today
// (inclusive of today as day 0) to the target date, ignoring clock time.
func calendarDaysFromToday(date time.Time) int {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return int(target.Sub(today).Hours() / 24)
}
func (u *RestockUseCase) resolveProducts(ctx context.Context, storeID string, productID string) ([]product_client.ProductDTO, error) {
	if productID == "" {
		products, err := u.ProductClient.ListByStoreID(ctx, storeID)
		if err != nil {
			u.Log.Errorf("Failed to get products: %+v", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data produk")
		}
		return products, nil
	}

	product, err := u.ProductClient.GetByID(ctx, productID)
	if err != nil {
		u.Log.Warnf("Product not found : %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Produk tidak ditemukan")
	}
	if product.StoreID != storeID {
		return nil, fiber.NewError(fiber.StatusForbidden, "Produk tidak dimiliki toko ini")
	}

	return []product_client.ProductDTO{*product}, nil
}

func (u *RestockUseCase) predictEligibleItems(ctx context.Context, req *model.GenerateRestockPredictionRequest, forecastDate string, eligible []productWithHistory, skipped *[]model.RestockSkippedItemResponse) []productPrediction {
	if len(eligible) == 0 {
		return nil
	}

	batchRequest := mlclient.InventoryPredictionBatchRequest{Predictions: make([]mlclient.InventoryPredictionRequest, len(eligible))}

	for i, item := range eligible {
		batchRequest.Predictions[i] = mlclient.InventoryPredictionRequest{
			Store:        req.StoreID,
			Item:         item.Product.ID,
			Date:         forecastDate,
			SalesHistory: item.History,
		}
	}

	batchResponse, err := u.MLClient.PredictInventoryBatch(ctx, batchRequest)
	if err != nil {
		u.Log.Warnf("Batch inventory prediction failed, falling back to per-product calls: %+v", err)
		return u.predictEligibleItemsIndividually(ctx, req, forecastDate, eligible, skipped)
	}

	u.Log.Infof("Response ml: %+v", batchResponse)

	responseByProductID := make(map[string]mlclient.InventoryPredictionResponse, len(batchResponse))

	for _, response := range batchResponse {
		responseByProductID[response.Item] = response
	}

	predictions := make([]productPrediction, 0, len(eligible))
	for _, item := range eligible {
		response, ok := responseByProductID[item.Product.ID]
		if !ok {
			*skipped = append(*skipped, skippedItem(item.Product, skipReasonMLError))
			continue
		}
		predictions = append(predictions, productPrediction{Product: item.Product, Response: response})
	}

	return predictions
}

func (u *RestockUseCase) predictEligibleItemsIndividually(ctx context.Context, req *model.GenerateRestockPredictionRequest, forecastDate string, eligible []productWithHistory, skipped *[]model.RestockSkippedItemResponse) []productPrediction {
	predictions := make([]productPrediction, 0, len(eligible))
	for _, item := range eligible {
		response, err := u.MLClient.PredictInventory(ctx, mlclient.InventoryPredictionRequest{
			Store:        req.StoreID,
			Item:         item.Product.ID,
			Date:         forecastDate,
			SalesHistory: item.History,
		})
		if err != nil {
			u.Log.Warnf("Inventory prediction failed for product %s: %+v", item.Product.ID, err)
			*skipped = append(*skipped, skippedItem(item.Product, skipReasonMLError))
			continue
		}
		predictions = append(predictions, productPrediction{Product: item.Product, Response: *response})
	}
	return predictions
}

func skippedItem(product product_client.ProductDTO, reason string) model.RestockSkippedItemResponse {
	return model.RestockSkippedItemResponse{
		ProductID:   product.ID,
		ProductName: product.ProductName,
		Reason:      reason,
	}
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}



type productWithHistory struct {
	Product product_client.ProductDTO
	History []float64
}

type productPrediction struct {
	Product  product_client.ProductDTO
	Response mlclient.InventoryPredictionResponse
}

func sumFloat(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}
