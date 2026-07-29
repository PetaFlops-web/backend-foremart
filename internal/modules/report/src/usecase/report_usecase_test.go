package usecase

import (
	"context"
	"testing"

	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	report_model "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/report/src/model"
	store_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store-client"
	transaction_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction-client"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type mockTransactionClient struct {
	items []transaction_client.TransactionItemDTO
}

func (m *mockTransactionClient) ListByStoreAndDate(ctx context.Context, storeId string, date string) ([]transaction_client.TransactionDTO, error) {
	return nil, nil
}

func (m *mockTransactionClient) ListItemsByStoreAndDate(ctx context.Context, storeId string, date string) ([]transaction_client.TransactionItemDTO, error) {
	return m.items, nil
}

func (m *mockTransactionClient) ListItemsByProduct(ctx context.Context, productId string, lookbackDays int) ([]transaction_client.TransactionItemDTO, error) {
	return nil, nil
}

func (m *mockTransactionClient) SumQtyByProductInMonth(ctx context.Context, productId string, yearMonth string) (int, error) {
	return 0, nil
}

type mockProductClient struct {
	products []product_client.ProductDTO
}

func (m *mockProductClient) GetByID(ctx context.Context, id string) (*product_client.ProductDTO, error) {
	return nil, nil
}

func (m *mockProductClient) ListByStoreID(ctx context.Context, storeId string) ([]product_client.ProductDTO, error) {
	return m.products, nil
}

func (m *mockProductClient) DecrementStock(ctx context.Context, id string, qty int) error {
	return nil
}

func (m *mockProductClient) IncrementStock(ctx context.Context, id string, qty int) error {
	return nil
}

func (m *mockProductClient) Search(ctx context.Context, storeId string, keyword string) ([]product_client.ProductDTO, error) {
	return nil, nil
}

type mockStoreClient struct {
	store *store_client.StoreDTO
}

func (m *mockStoreClient) GetStoreByID(ctx context.Context, storeID string) (*store_client.StoreDTO, error) {
	return m.store, nil
}

func (m *mockStoreClient) GetStoreByUserID(ctx context.Context, userID string) (*store_client.StoreDTO, error) {
	return m.store, nil
}

func TestReportUseCaseDaily(t *testing.T) {
	useCase := NewReportUseCase(
		logrus.New(),
		validator.New(),
		&mockTransactionClient{
			items: []transaction_client.TransactionItemDTO{
				{
					TransactionID:        "txn-1",
					ProductID:            "prod-1",
					ProductNameSnapshot:  "Beras",
					Qty:                  2,
					CostPriceSnapshot:    10000,
					SellingPriceSnapshot: 15000,
				},
				{
					TransactionID:        "txn-1",
					ProductID:            "prod-2",
					ProductNameSnapshot:  "Gula",
					Qty:                  1,
					CostPriceSnapshot:    12000,
					SellingPriceSnapshot: 17000,
				},
				{
					TransactionID:        "txn-2",
					ProductID:            "prod-1",
					ProductNameSnapshot:  "Beras",
					Qty:                  3,
					CostPriceSnapshot:    10000,
					SellingPriceSnapshot: 15000,
				},
			},
		},
		&mockProductClient{
			products: []product_client.ProductDTO{
				{ID: "prod-1", ProductName: "Beras", Stock: 10, Unit: "kg"},
				{ID: "prod-2", ProductName: "Gula", Stock: 8, Unit: "kg"},
			},
		},
		&mockStoreClient{
			store: &store_client.StoreDTO{
				ID:        "store-1",
				UserID:    "user-1",
				StoreName: "Toko Maju",
			},
		},
	)

	response, err := useCase.Daily(context.Background(), &report_model.DailyReportRequest{
		StoreID: "store-1",
		Date:    "2026-07-29",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.TotalOmset != 92000 {
		t.Fatalf("expected total omset 92000, got %d", response.TotalOmset)
	}

	if response.TotalUntung != 30000 {
		t.Fatalf("expected total untung 30000, got %d", response.TotalUntung)
	}

	if response.JumlahTransaksi != 2 {
		t.Fatalf("expected jumlah transaksi 2, got %d", response.JumlahTransaksi)
	}

	if len(response.ProdukTerlaris) != 2 {
		t.Fatalf("expected 2 best selling products, got %d", len(response.ProdukTerlaris))
	}

	if response.ProdukTerlaris[0].ProductID != "prod-1" || response.ProdukTerlaris[0].QtySold != 5 {
		t.Fatalf("unexpected first best seller: %+v", response.ProdukTerlaris[0])
	}

	if len(response.SisaStok) != 2 {
		t.Fatalf("expected 2 stock summaries, got %d", len(response.SisaStok))
	}

	if response.Store.StoreName != "Toko Maju" {
		t.Fatalf("expected store name Toko Maju, got %s", response.Store.StoreName)
	}
}
