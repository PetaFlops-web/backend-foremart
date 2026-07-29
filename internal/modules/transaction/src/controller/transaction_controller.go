package controller

import (
	"bytes"
	"io"
	"math"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type TransactionController struct {
	Log     *logrus.Logger
	UseCase *usecase.TransactionUseCase
}

func NewTransactionController(useCase *usecase.TransactionUseCase, logger *logrus.Logger) *TransactionController {
	return &TransactionController{
		Log:     logger,
		UseCase: useCase,
	}
}

// ExtractVoice godoc
// @Summary      Ekstrak transaksi dari suara
// @Description  Mengirim file audio ke ML untuk ditranskripsi, lalu mencocokkan dengan produk di DB untuk direview oleh user.
// @Tags         Transaction
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        store_id formData string true "Store ID"
// @Param        file     formData file   true "Audio File (.wav, .mp3, dll)"
// @Success      200      {object} response.WebResponse[model.TransactionPreviewResponse]
// @Failure      400      {object} response.ApiErrorResponse
// @Failure      500      {object} response.ApiErrorResponse
// @Router       /transactions/extract/voice [post]
func (c *TransactionController) ExtractVoice(ctx *fiber.Ctx) error {
	storeId := ctx.FormValue("store_id")
	if storeId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "store_id wajib diisi")
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file audio tidak ditemukan")
	}

	fileData, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal membaca file audio")
	}
	defer fileData.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, fileData); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal memproses isi file audio")
	}

	request := &model.ExtractVoiceRequest{
		StoreId:   storeId,
		AudioData: buf.Bytes(),
		Filename:  file.Filename,
	}

	resp, err := c.UseCase.ExtractVoice(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[*model.TransactionPreviewResponse]{
		Data:    resp,
		Message: "Berhasil mengekstrak transaksi dari suara",
		Success: true,
	})
}

// Create godoc
// @Summary      Konfirmasi & simpan transaksi
// @Description  Menyimpan transaksi yang telah direview oleh user dan mengurangi stok produk
// @Tags         Transaction
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body     model.CreateTransactionRequest true "Data Transaksi Final"
// @Success      200  {object} response.WebResponse[model.TransactionResponse]
// @Failure      400  {object} response.ApiErrorResponse
// @Failure      500  {object} response.ApiErrorResponse
// @Router       /transactions [post]
func (c *TransactionController) Create(ctx *fiber.Ctx) error {
	request := new(model.CreateTransactionRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	resp, err := c.UseCase.Create(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[*model.TransactionResponse]{
		Data:    resp,
		Message: "Berhasil menyimpan transaksi",
		Success: true,
	})
}

// Get godoc
// @Summary      Menampilkan detail satu transaksi
// @Description  Mengambil data transaksi beserta daftar item-nya
// @Tags         Transaction
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path     string true "Transaction ID"
// @Param        store_id query    string true "Store ID"
// @Success      200      {object} response.WebResponse[model.TransactionResponse]
// @Failure      400      {object} response.ApiErrorResponse
// @Failure      404      {object} response.ApiErrorResponse
// @Router       /transactions/{id} [get]
func (c *TransactionController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	storeId := ctx.Query("store_id")

	if storeId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Store ID tidak boleh kosong")
	}

	resp, err := c.UseCase.Get(ctx.UserContext(), storeId, id)
	if err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[*model.TransactionResponse]{
		Data:    resp,
		Message: "Berhasil mengambil data transaksi",
		Success: true,
	})
}

// Delete godoc
// @Summary      Menghapus transaksi
// @Description  Menghapus data transaksi dan mengembalikan stok produk (increment)
// @Tags         Transaction
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path     string true "Transaction ID"
// @Param        store_id query    string true "Store ID"
// @Success      200      {object} response.WebResponse[any]
// @Failure      400      {object} response.ApiErrorResponse
// @Failure      404      {object} response.ApiErrorResponse
// @Router       /transactions/{id} [delete]
func (c *TransactionController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	storeId := ctx.Query("store_id")

	if storeId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Store ID tidak boleh kosong")
	}

	if err := c.UseCase.Delete(ctx.UserContext(), storeId, id); err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[any]{
		Data:    nil,
		Message: "Berhasil menghapus transaksi",
		Success: true,
	})
}

// List godoc
// @Summary      Menampilkan riwayat transaksi
// @Description  Menampilkan daftar transaksi toko dengan pagination
// @Tags         Transaction
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        store_id query    string true  "Store ID"
// @Param        page     query    int    false "Nomor Halaman" default(1)
// @Param        size     query    int    false "Ukuran Halaman" default(10)
// @Success      200      {object} response.WebResponse[[]model.TransactionResponse]
// @Failure      400      {object} response.ApiErrorResponse
// @Failure      500      {object} response.ApiErrorResponse
// @Router       /transactions [get]
func (c *TransactionController) List(ctx *fiber.Ctx) error {
	request := &model.SearchTransactionRequest{
		StoreId: ctx.Query("store_id", ""),
		Page:    ctx.QueryInt("page", 1),
		Size:    ctx.QueryInt("size", 10),
	}

	responses, total, err := c.UseCase.List(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	paging := &response.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return ctx.JSON(response.WebResponse[[]model.TransactionResponse]{
		Data:    responses,
		Paging:  paging,
		Message: "Berhasil menampilkan riwayat transaksi",
		Success: true,
	})
}