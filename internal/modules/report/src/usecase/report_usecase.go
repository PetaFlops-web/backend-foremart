package usecase

import (
	"context"
	"sort"
	"time"

	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	report_model "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/report/src/model"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ReportUseCase struct {
	Log               *logrus.Logger
	Validate          *validator.Validate
	TransactionClient transaction_client.Client
	ProductClient     product_client.Client
	StoreClient       store_client.StoreClient
}

func NewReportUseCase(
	log *logrus.Logger,
	validate *validator.Validate,
	transactionClient transaction_client.Client,
	productClient product_client.Client,
	storeClient store_client.StoreClient,
) *ReportUseCase {
	return &ReportUseCase{
		Log:               log,
		Validate:          validate,
		TransactionClient: transactionClient,
		ProductClient:     productClient,
		StoreClient:       storeClient,
	}
}

func (u *ReportUseCase) Daily(ctx context.Context, request *report_model.DailyReportRequest) (*report_model.DailyReportResponse, error) {
	if err := u.Validate.Struct(request); err != nil {
		u.Log.Warnf("Invalid report request : %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Format permintaan laporan tidak valid")
	}

	reportDate := time.Now().Format("2006-01-02")
	if request.Date != "" {
		if _, err := time.Parse("2006-01-02", request.Date); err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Format tanggal tidak valid, gunakan YYYY-MM-DD")
		}
		reportDate = request.Date
	}

	store, err := u.StoreClient.GetStoreByID(ctx, request.StoreID)
	if err != nil {
		u.Log.Warnf("Store not found : %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Store tidak ditemukan")
	}

	items, err := u.TransactionClient.ListItemsByStoreAndDate(ctx, request.StoreID, reportDate)
	if err != nil {
		u.Log.Errorf("Failed to get transaction items : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data item transaksi")
	}

	products, err := u.ProductClient.ListByStoreID(ctx, request.StoreID)
	if err != nil {
		u.Log.Errorf("Failed to get products : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data stok produk")
	}

	productMap := make(map[string]product_client.ProductDTO, len(products))
	stockSummaries := make([]report_model.StockSummaryResponse, 0, len(products))
	for _, product := range products {
		productMap[product.ID] = product
		stockSummaries = append(stockSummaries, report_model.StockSummaryResponse{
			ProductID:   product.ID,
			ProductName: product.ProductName,
			Stock:       product.Stock,
			Unit:        product.Unit,
		})
	}

	sort.Slice(stockSummaries, func(i, j int) bool {
		return stockSummaries[i].ProductName < stockSummaries[j].ProductName
	})

	totalOmset := int64(0)
	totalUntung := int64(0)
	transactionIDs := make(map[string]struct{})

	type bestsellerAggregate struct {
		ProductID   string
		ProductName string
		QtySold     int
	}

	bestSellerMap := make(map[string]*bestsellerAggregate)

	for _, item := range items {
		totalOmset += int64(item.Qty) * item.SellingPriceSnapshot
		totalUntung += int64(item.Qty) * (item.SellingPriceSnapshot - item.CostPriceSnapshot)
		transactionIDs[item.TransactionID] = struct{}{}

		aggregate, exists := bestSellerMap[item.ProductID]
		if !exists {
			aggregate = &bestsellerAggregate{
				ProductID:   item.ProductID,
				ProductName: item.ProductNameSnapshot,
			}
			bestSellerMap[item.ProductID] = aggregate
		}
		aggregate.QtySold += item.Qty
	}

	bestSellers := make([]report_model.BestSellingProductResponse, 0, len(bestSellerMap))
	for _, aggregate := range bestSellerMap {
		product := productMap[aggregate.ProductID]
		bestSellers = append(bestSellers, report_model.BestSellingProductResponse{
			ProductID:    aggregate.ProductID,
			ProductName:  aggregate.ProductName,
			QtySold:      aggregate.QtySold,
			CurrentStock: product.Stock,
			Unit:         product.Unit,
		})
	}

	sort.Slice(bestSellers, func(i, j int) bool {
		if bestSellers[i].QtySold == bestSellers[j].QtySold {
			return bestSellers[i].ProductName < bestSellers[j].ProductName
		}
		return bestSellers[i].QtySold > bestSellers[j].QtySold
	})

	if len(bestSellers) > 5 {
		bestSellers = bestSellers[:5]
	}

	return &report_model.DailyReportResponse{
		Store: report_model.StoreSummaryResponse{
			ID:        store.ID,
			UserID:    store.UserID,
			StoreName: store.StoreName,
		},
		Date:            reportDate,
		TotalOmset:      totalOmset,
		TotalUntung:     totalUntung,
		JumlahTransaksi: len(transactionIDs),
		ProdukTerlaris:  bestSellers,
		SisaStok:        stockSummaries,
	}, nil
}
