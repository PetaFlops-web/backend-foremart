package usecase

import (
	"context"
	"time"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product-client"
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
}

func NewTransactionUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	transactionRepo *repository.TransactionRepository,
	transactionItemRepo *repository.TransactionItemRepository,
	productClient product_client.Client,
	mlClient mlclient.MLClient,
) *TransactionUseCase {
	return &TransactionUseCase{
		DB:                  db,
		Log:                 log,
		Validate:            validate,
		TransactionRepo:     transactionRepo,
		TransactionItemRepo: transactionItemRepo,
		ProductClient:       productClient,
		MLClient:            mlClient,
	}
}

// ExtractVoice handles audio transcription and transaction preview merging.
func (u *TransactionUseCase) ExtractVoice(ctx context.Context, req *model.ExtractVoiceRequest) (*model.TransactionPreviewResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.Warnf("Invalid extract request: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// 1. Call ML Service
	mlRes, err := u.MLClient.TranscribeAndExtract(ctx, req.AudioData, req.Filename)
	if err != nil {
		u.Log.Errorf("Failed calling ML service: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Gagal memproses audio transaksi")
	}

	// 2. Build Preview Items
	previewResponse := &model.TransactionPreviewResponse{
		RawText: mlRes.RawText,
		Items:   make([]model.TransactionPreviewItemResponse, len(mlRes.Items)),
	}

	for i, extractedItem := range mlRes.Items {
		previewItem := model.TransactionPreviewItemResponse{
			RawText:       extractedItem.Item,
			DetectedQty:   extractedItem.Qty,
			DetectedPrice: extractedItem.Harga,
			IsMatched:     false,
		}

		// Stage 1: Search using produk_katalog (ML standard)
		var bestMatch *product_client.ProductDTO
		if extractedItem.ProdukKatalog != "" {
			results, err := u.ProductClient.Search(ctx, req.StoreId, extractedItem.ProdukKatalog)
			if err == nil && len(results) > 0 {
				bestMatch = &results[0]
			}
		}

		// Stage 2: Fallback search using raw item
		if bestMatch == nil && extractedItem.Item != "" {
			results, err := u.ProductClient.Search(ctx, req.StoreId, extractedItem.Item)
			if err == nil && len(results) > 0 {
				bestMatch = &results[0]
			}
		}

		// Populate DB data if matched
		if bestMatch != nil {
			previewItem.IsMatched = true
			previewItem.ProductId = bestMatch.ID
			previewItem.ProductName = bestMatch.ProductName
			previewItem.SellingPrice = bestMatch.SellingPrice
			previewItem.CostPrice = bestMatch.CostPrice
			previewItem.Stock = bestMatch.Stock
		}

		previewResponse.Items[i] = previewItem
	}

	return previewResponse, nil
}

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

	txn := &entity.Transaction{
		ID:              txnID,
		StoreID:         req.StoreId,
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