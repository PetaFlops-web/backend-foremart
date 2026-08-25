package controller

import (
	"math"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/customer/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type CustomerController struct {
	Log     *logrus.Logger
	UseCase *usecase.CustomerUseCase
}

func NewCustomerController(useCase *usecase.CustomerUseCase, logger *logrus.Logger) *CustomerController {
	return &CustomerController{
		Log:     logger,
		UseCase: useCase,
	}
}

// Create godoc
// @Summary      Menambahkan data customer baru
// @Description  Membuat entri customer baru atau mengembalikan yang sudah ada (upsert by phone)
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateCustomerRequest  true  "Data Customer"
// @Success      200   {object}  response.WebResponse[model.CustomerResponse]
// @Failure      400   {object}  response.ApiErrorResponse
// @Failure      500   {object}  response.ApiErrorResponse
// @Router       /customers [post]
func (c *CustomerController) Create(ctx *fiber.Ctx) error {
	request := new(model.CreateCustomerRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse request body : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	resp, err := c.UseCase.Create(ctx.UserContext(), request)
	if err != nil {
		c.Log.Warnf("Failed to create customer : %+v", err)
		return err
	}

	return ctx.JSON(response.WebResponse[*model.CustomerResponse]{
		Data:    resp,
		Message: "Berhasil menambahkan customer",
		Success: true,
	})
}

// Get godoc
// @Summary      Menampilkan detail satu customer
// @Description  Mengambil data customer berdasarkan ID dan Store ID
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Customer ID"
// @Param        store_id query     string  true  "Store ID"
// @Success      200      {object}  response.WebResponse[model.CustomerResponse]
// @Failure      400      {object}  response.ApiErrorResponse
// @Failure      404      {object}  response.ApiErrorResponse
// @Router       /customers/{id} [get]
func (c *CustomerController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Customer ID tidak valid")
	}
	storeID := ctx.Query("store_id", "")

	if storeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Store ID tidak boleh kosong")
	}

	resp, err := c.UseCase.Get(ctx.UserContext(), storeID, id)
	if err != nil {
		c.Log.Warnf("Failed to get customer : %+v", err)
		return err
	}

	return ctx.JSON(response.WebResponse[*model.CustomerResponse]{
		Data:    resp,
		Message: "Berhasil mengambil data customer",
		Success: true,
	})
}

// Search godoc
// @Summary      Menampilkan daftar customer dengan pagination
// @Description  Mencari dan menampilkan daftar customer untuk suatu toko (scoped per-toko)
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        store_id query     string  true   "Store ID"
// @Param        search   query     string  false  "Keyword nama atau nomor telepon"
// @Param        page     query     int     false  "Nomor Halaman" default(1)
// @Param        size     query     int     false  "Ukuran Halaman" default(10)
// @Success      200      {object}  response.WebResponse[[]model.CustomerResponse]
// @Failure      400      {object}  response.ApiErrorResponse
// @Failure      500      {object}  response.ApiErrorResponse
// @Router       /customers [get]
func (c *CustomerController) Search(ctx *fiber.Ctx) error {
	request := &model.SearchCustomerRequest{
		StoreID: ctx.Query("store_id", ""),
		Search:  ctx.Query("search", ""),
		Page:    ctx.QueryInt("page", 1),
		Size:    ctx.QueryInt("size", 10),
	}

	responses, total, err := c.UseCase.Search(ctx.UserContext(), request)
	if err != nil {
		c.Log.Warnf("Failed searching customer : %+v", err)
		return err
	}

	paging := &response.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return ctx.JSON(response.WebResponse[[]model.CustomerResponse]{
		Data:    responses,
		Paging:  paging,
		Message: "Berhasil menampilkan daftar customer",
		Success: true,
	})
}
