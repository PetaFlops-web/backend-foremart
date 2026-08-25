package usecase

import (
	"context"
	"time"

	customer_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer-client"
	product_client "github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/model/converter"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/repository"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/mlclient"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionUseCase struct {
	DB                      *gorm.DB
	Log                     *logrus.Logger
	Validate                *validator.Validate
	TransactionRepo         *repository.TransactionRepository
	TransactionItemRepo     *repository.TransactionItemRepository
	ProductClient           product_client.Client
	MLClient                mlclient.MLClient
	CustomerClient          customer_client.Client
}

func NewTransactionUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	transactionRepo *repository.TransactionRepository,
	transactionItemRepo *repository.TransactionItemRepository,
	productClient product_client.Client,
	mlClient mlclient.MLClient,
	customerClient customer_client.Client,
) *TransactionUseCase {
	return &TransactionUseCase{
		DB:                  db,
		Log:                 log,
		Validate:            validate,
		TransactionRepo:     transactionRepo,
		TransactionItemRepo: transactionItemRepo,
		ProductClient:       productClient,
		MLClient:            mlClient,
		CustomerClient:      customerClient,
	}
}

// ExtractVoice handles audio transcription and transaction preview merging.
// func (u *TransactionUseCase) ExtractVoice(ctx context.Context, req *model.ExtractVoiceRequest) (*model.TransactionPreviewResponse, error) {
// 	if err := u.Validate.Struct(req); err != nil {
// 		u.Log.Warnf("Invalid extract request: %+v", err)
// 		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
// 	}

// 	// 1. Call ML Service
// 	mlRes, err := u.MLClient.TranscribeAndExtract(ctx, req.AudioData, req.Filename)
// 	if err != nil {
// 		u.Log.Errorf("Failed calling ML service: %+v", err)
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal memproses audio transaksi")
// 	}

// 	// 2. Build Preview Items
// 	previewResponse := &model.TransactionPreviewResponse{
// 		RawText: mlRes.RawText,
// 		Items:   make([]model.TransactionPreviewItemResponse, len(mlRes.Items)),
// 	}

// 	for i, extractedItem := range mlRes.Items {
// 		previewItem := model.TransactionPreviewItemResponse{
// 			RawText:       extractedItem.Item,
// 			DetectedQty:   extractedItem.Qty,
// 			DetectedPrice: extractedItem.Harga,
// 			IsMatched:     false,
// 		}

// 		// Stage 1: Search using produk_katalog (ML standard)
// 		var bestMatch *product_client.ProductDTO
// 		if extractedItem.ProdukKatalog != "" {
// 			results, err := u.ProductClient.Search(ctx, req.StoreId, extractedItem.ProdukKatalog)
// 			if err == nil && len(results) > 0 {
// 				bestMatch = &results[0]
// 			}
// 		}

// 		// Stage 2: Fallback search using raw item
// 		if bestMatch == nil && extractedItem.Item != "" {
// 			results, err := u.ProductClient.Search(ctx, req.StoreId, extractedItem.Item)
// 			if err == nil && len(results) > 0 {
// 				bestMatch = &results[0]
// 			}
// 		}

// 		// Populate DB data if matched
// 		if bestMatch != nil {
// 			previewItem.IsMatched = true
// 			previewItem.ProductId = bestMatch.ID
// 			previewItem.ProductName = bestMatch.ProductName
// 			previewItem.SellingPrice = bestMatch.SellingPrice
// 			previewItem.CostPrice = bestMatch.CostPrice
// 			previewItem.Stock = bestMatch.Stock
// 		}

// 		previewResponse.Items[i] = previewItem
// 	}

// 	return previewResponse, nil
// }

// Create persists a confirmed transaction into the database.
func (u *TransactionUseCase) Create(ctx context.Context, req *model.CreateTransactionRequest) (*model.TransactionResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.Warnf("Invalid create request: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Data konfirmasi transaksi tidak valid")
	}

	if len(req.Items) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Transaksi tidak boleh kosong")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// 1. Setup Transaction entity
	txnID, err := utils.GenerateTransactionId()
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat ID Transaksi")
	}

	// Validate customer via customer-client
	customer, err := u.CustomerClient.GetByID(ctx, req.CustomerID)
	if err != nil {
		u.Log.Warnf("Customer %s not found: %v", req.CustomerID, err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Customer tidak ditemukan")
	}

	txn := &entity.Transaction{
		ID:              txnID,
		StoreID:         req.StoreId,
		CustomerID:      req.CustomerID,
		CustomerPhone:   customer.Phone,
		CustomerName:    customer.Name,
		TransactionDate: time.Now(),
		Source:          req.Source,
	}

	if err := u.TransactionRepo.Create(tx, txn); err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan transaksi")
	}

	// 2. Process Items and update Stock
	var txnItems []entity.TransactionItem
	for _, itemReq := range req.Items {
		// Verify product
		product, err := u.ProductClient.GetByID(ctx, itemReq.ProductId)
		if err != nil {
			u.Log.Warnf("Product %s not found", itemReq.ProductId)
			return nil, fiber.NewError(fiber.StatusNotFound, "Produk tidak ditemukan")
		}

		if product.StoreID != req.StoreId {
			return nil, fiber.NewError(fiber.StatusForbidden, "Produk tidak termasuk dalam toko Anda")
		}

		// Check stock
		if product.Stock < itemReq.Qty {
			u.Log.Warnf("Insufficient stock for %s. Have: %d, Need: %d", product.ProductName, product.Stock, itemReq.Qty)
			return nil, fiber.NewError(fiber.StatusBadRequest, "Stok produk tidak mencukupi")
		}

		// Decrement stock
		if err := u.ProductClient.DecrementStock(ctx, itemReq.ProductId, itemReq.Qty); err != nil {
			u.Log.Errorf("Failed to decrement stock for %s: %v", itemReq.ProductId, err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengurangi stok produk")
		}

		// Create snapshot
		itemID, _ := utils.GenerateTransactionItemId()
		txnItems = append(txnItems, entity.TransactionItem{
			ID:                   itemID,
			TransactionID:        txnID,
			ProductID:            product.ID,
			ProductNameSnapshot:  product.ProductName,
			Qty:                  itemReq.Qty,
			CostPriceSnapshot:    product.CostPrice,
			SellingPriceSnapshot: itemReq.SellingPriceFinal, // Final price from confirmation
		})
	}

	// 3. Save items
	if err := u.TransactionItemRepo.CreateBatch(tx, txnItems); err != nil {
		u.Log.Errorf("Failed to save transaction items: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan detail transaksi")
	}

	// 4. Commit
	if err := tx.Commit().Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyelesaikan transaksi")
	}

	txn.Items = txnItems
	return converter.TransactionToResponse(txn), nil
}

// Get retrieves a specific transaction by ID.
func (u *TransactionUseCase) Get(ctx context.Context, storeId string, id string) (*model.TransactionResponse, error) {
	txn, err := u.TransactionRepo.FindByIdWithItems(u.DB.WithContext(ctx), id)
	if err != nil {
		u.Log.Warnf("Transaction not found: %v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Transaksi tidak ditemukan")
	}

	if txn.StoreID != storeId {
		return nil, fiber.NewError(fiber.StatusForbidden, "Akses ditolak")
	}

	return converter.TransactionToResponse(txn), nil
}

// Delete removes a transaction and restores product stock.
func (u *TransactionUseCase) Delete(ctx context.Context, storeId string, id string) error {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// 1. Get transaction with items
	txn, err := u.TransactionRepo.FindByIdWithItems(tx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Transaksi tidak ditemukan")
	}

	if txn.StoreID != storeId {
		return fiber.NewError(fiber.StatusForbidden, "Akses ditolak")
	}

	// 2. Restore stock for each item
	for _, item := range txn.Items {
		if err := u.ProductClient.IncrementStock(ctx, item.ProductID, item.Qty); err != nil {
			u.Log.Errorf("Failed restoring stock for product %s: %v", item.ProductID, err)
			return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengembalikan stok produk")
		}
	}

	// 3. Delete items
	if err := u.TransactionItemRepo.DeleteByTransactionId(tx, txn.ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menghapus detail transaksi")
	}

	// 4. Delete transaction
	if err := u.TransactionRepo.Delete(tx, txn); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menghapus transaksi")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menyelesaikan penghapusan transaksi")
	}

	return nil
}

// List retrieves paginated transactions for a store.
func (u *TransactionUseCase) List(ctx context.Context, req *model.SearchTransactionRequest) ([]model.TransactionResponse, int64, error) {
	if err := u.Validate.Struct(req); err != nil {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, "Format permintaan tidak valid")
	}

	txns, total, err := u.TransactionRepo.FindByStoreId(u.DB.WithContext(ctx), req.StoreId, req.Page, req.Size)
	if err != nil {
		u.Log.Errorf("Failed to list transactions: %v", err)
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data transaksi")
	}

	responses := make([]model.TransactionResponse, len(txns))
	for i, txn := range txns {
		responses[i] = *converter.TransactionToResponse(&txn)
	}

	return responses, total, nil
}

// SeedMock generates backdated mock transactions so the ML pipeline has enough
// history to produce a prediction. It is a development/demo helper and is
// intentionally separate from the real Create() flow so it can stamp arbitrary
// transaction dates (Create() always uses time.Now()).
//
// Two modes:
//   - "restock": backfills ~30 days of daily sales for every product so the
//     restock prediction (/restock-predictions/_generate) has a non-zero
//     history to forecast from.
//   - "survival": creates 3 purchase events for each customer×product pair,
//     spaced so the survival prediction lands within the reminder threshold
//     (DaysThreshold = 3), so the notification job will actually send.
func (u *TransactionUseCase) SeedMock(ctx context.Context, req *model.SeedMockRequest) (*model.SeedMockResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.Warnf("Invalid mock seed request: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Data request mock tidak valid")
	}

	products, err := u.ProductClient.ListByStoreID(ctx, req.StoreID)
	if err != nil {
		u.Log.Errorf("Failed to list products for mock seed: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data produk")
	}
	if len(products) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Tidak ada produk untuk di-mock. Tambahkan produk dulu.")
	}

	switch req.Mode {
	case "restock":
		return u.seedRestockHistory(ctx, req.StoreID, products)
	case "survival":
		return u.seedSurvivalHistory(ctx, req.StoreID, products)
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Mode mock tidak dikenal")
	}
}

// seedRestockHistory creates daily sales for the last 35 days across all
// products with more realistic gaps between transactions. Instead of every day,
// we simulate real purchasing behavior where some days have multiple sales,
// some days are skipped. This produces varied patterns that ML can learn from.
func (u *TransactionUseCase) seedRestockHistory(ctx context.Context, storeID string, products []product_client.ProductDTO) (*model.SeedMockResponse, error) {
	now := time.Now()
	totalCreated := 0

	// Realistic transaction dates with gaps (simulating real buying patterns)
	// Not every single day, but clustered around certain periods
	dateOffsets := []int{
		35, 34, 33,      // Day 1-3: active period
		32, 31,          // gap
		30, 29, 28, 27, 26, 25, // Day 2-week: moderate activity
		24, 23, 22,     // slight dip
		21, 20,         // weekend-like cluster
		19, 18, 17, 16, // recovery
		15, 14,         // quiet
		13, 12, 11, 10, // active again
		9, 8,           // dip
		7, 6, 5,        // cluster
		4, 3, 2, 1,     // recent activity
	}

	for _, dayOffset := range dateOffsets {
		date := now.AddDate(0, 0, -dayOffset)

		// Decide which products get sold on this day (realistic variation)
		var productsToSell []int
		switch dayOffset {
		case 33, 28, 22, 16, 10, 5: // High-activity days: both products sell (if available)
			if len(products) >= 2 {
				productsToSell = []int{0, 1}
			} else {
				productsToSell = []int{0}
			}
		case 32, 31, 21, 20, 15, 14, 7, 6, 3, 2: // Single product sells
			productsToSell = []int{dayOffset % len(products)}
		default: // No sales on most days (realistic restocking pattern)
			continue
		}

		for _, prodIdx := range productsToSell {
			product := products[prodIdx]

			// Higher qty on high-activity days to deplete stock realistically
			var qty int
			if prodIdx == 0 { // minyak (index 0) - lower original stock means less needed depletion
				qty = 12 + ((dayOffset%5)*3) // Very aggressive sales for minyak
			} else { // beras (index 1) - even higher to deplete faster
				qty = 20 + ((dayOffset%5)*4) // Maximize sales
			}

			txnID, err := utils.GenerateTransactionId()
			if err != nil {
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat ID transaksi")
			}
			itemID, err := utils.GenerateTransactionItemId()
			if err != nil {
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat ID item transaksi")
			}

			txn := &entity.Transaction{
				ID:              txnID,
				StoreID:         storeID,
				CustomerID:      "",
				CustomerPhone:   "",
				CustomerName:    "",
				TransactionDate: date,
				Source:          "mock",
			}
			item := entity.TransactionItem{
				ID:                   itemID,
				TransactionID:        txnID,
				ProductID:            product.ID,
				ProductNameSnapshot:  product.ProductName,
				Qty:                  qty,
				CostPriceSnapshot:    product.CostPrice,
				SellingPriceSnapshot: product.SellingPrice,
			}

			// Decrement stock to reduce inventory realistically
			if err := u.ProductClient.DecrementStock(ctx, product.ID, qty); err != nil {
				u.Log.Errorf("Failed to decrement stock for %s during mock: %v", product.ProductName, err)
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengurangi stok produk selama mock")
			}

			if err := u.TransactionRepo.Create(u.DB.WithContext(ctx), txn); err != nil {
				u.Log.Errorf("Failed to create mock transaction: %v", err)
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan transaksi mock")
			}
			if err := u.TransactionItemRepo.Create(u.DB.WithContext(ctx), &item); err != nil {
				u.Log.Errorf("Failed to create mock transaction item: %v", err)
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan item transaksi mock")
			}
			totalCreated++
		}
	}

	return &model.SeedMockResponse{
		Mode:                "restock",
		TransactionsCreated: totalCreated,
		ProductsAffected:    len(products),
		CustomersAffected:   0,
	}, nil
}

// seedSurvivalHistory creates 3 purchase events per customer×product pair. The
// latest purchase is placed ~1 day ago and earlier ones ~30 and ~60 days ago,
// so the survival model predicts a median survival around 30 days with only 1
// day elapsed — putting pred_days_left inside the reminder threshold.
func (u *TransactionUseCase) seedSurvivalHistory(ctx context.Context, storeID string, products []product_client.ProductDTO) (*model.SeedMockResponse, error) {
	now := time.Now()

	// Gather customers for this store. If none exist, mock will skip (require manual setup).
	customers, err := u.CustomerClient.ListByStoreID(ctx, storeID)
	if err != nil {
		u.Log.Errorf("Failed to list customers for mock seed: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data pelanggan")
	}

	if len(customers) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Tidak ada pelanggan untuk di-mock di toko ini.")
	}

	// Delete all existing transactions in this store so our backdated dates
	// are always the most recent and guarantee survival prediction triggers.
	if err := u.deleteStoreTransactions(ctx, storeID); err != nil {
		u.Log.Errorf("Failed to clear old store transactions: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menghapus transaksi lama")
	}

	// Last purchase should be close enough that pred_days_left drops below
	// threshold (≤ 3 days). We space purchases so:
	// - Latest ~1 day ago → model sees recent activity
	// - Earlier ~45 days ago → median survival computed correctly  
	// With stock codes, this gives pred_days_left ≈ 2-3 days when called today.
	// Last purchase >= 95 days ago ensures pred_days_left = 0 when median_survival is capped at 90
	// With median_survival capped at 90: pred_days_left = max(0, 90-95) = 0 ≤ threshold ✅
	spacingDays := []int{180, 120, 95} 
	totalCreated := 0

	for _, customer := range customers {
		for _, product := range products {
			for _, daysAgo := range spacingDays {
				date := now.AddDate(0, 0, -daysAgo)

				txnID, err := utils.GenerateTransactionId()
				if err != nil {
					return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat ID transaksi")
				}
				itemID, err := utils.GenerateTransactionItemId()
				if err != nil {
					return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat ID item transaksi")
				}

				txn := &entity.Transaction{
					ID:              txnID,
					StoreID:         storeID,
					CustomerID:      customer.ID,
					CustomerPhone:   customer.Phone,
					CustomerName:    customer.Name,
					TransactionDate: date,
					Source:          "mock",
				}
				item := entity.TransactionItem{
					ID:                   itemID,
					TransactionID:        txnID,
					ProductID:            product.ID,
					ProductNameSnapshot:  product.ProductName,
					Qty:                  1,
					CostPriceSnapshot:    product.CostPrice,
					SellingPriceSnapshot: product.SellingPrice,
				}

				if err := u.TransactionRepo.Create(u.DB.WithContext(ctx), txn); err != nil {
					u.Log.Errorf("Failed to create mock transaction: %v", err)
					return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan transaksi mock")
				}
				if err := u.TransactionItemRepo.Create(u.DB.WithContext(ctx), &item); err != nil {
					u.Log.Errorf("Failed to create mock transaction item: %v", err)
					return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan item transaksi mock")
				}
				totalCreated++
			}
		}
	}

	return &model.SeedMockResponse{
		Mode:                "survival",
		TransactionsCreated: totalCreated,
		ProductsAffected:    len(products),
		CustomersAffected:   len(customers),
	}, nil
}

// deleteStoreTransactions removes every transaction (and its items) for a
// store. Used by the mock seed to reset history so backdated mock dates are
// always the most recent.
func (u *TransactionUseCase) deleteStoreTransactions(ctx context.Context, storeID string) error {
	db := u.DB.WithContext(ctx)
	var ids []string
	if err := db.Model(&entity.Transaction{}).Where("store_id = ?", storeID).Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) > 0 {
		if err := db.Where("transaction_id IN ?", ids).Delete(&entity.TransactionItem{}).Error; err != nil {
			return err
		}
	}
	return db.Where("store_id = ?", storeID).Delete(&entity.Transaction{}).Error
}