package controller

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type RestockController struct {
	Log     *logrus.Logger
	UseCase *usecase.RestockUseCase
}

func NewRestockController(useCase *usecase.RestockUseCase, logger *logrus.Logger) *RestockController {
	return &RestockController{
		Log:     logger,
		UseCase: useCase,
	}
}

// Generate godoc
// @Summary      Generate prediksi restock
// @Description  Mengambil histori transaksi, memanggil ML inventory prediction, lalu menyimpan rekomendasi restock toko
// @Tags         Restock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body     model.GenerateRestockPredictionRequest true "Parameter prediksi restock"
// @Success      200  {object} response.WebResponse[model.GenerateRestockPredictionResponse]
// @Failure      400  {object} response.ApiErrorResponse
// @Failure      403  {object} response.ApiErrorResponse
// @Failure      404  {object} response.ApiErrorResponse
// @Failure      500  {object} response.ApiErrorResponse
// @Router       /restock-predictions/_generate [post]
func (c *RestockController) Generate(ctx *fiber.Ctx) error {
	request := new(model.GenerateRestockPredictionRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	resp, err := c.UseCase.Generate(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[*model.GenerateRestockPredictionResponse]{
		Data:    resp,
		Message: "Berhasil membuat prediksi restock",
		Success: true,
	})
}

// List godoc
// @Summary      Menampilkan prediksi restock
// @Description  Menampilkan prediksi restock tersimpan untuk toko
// @Tags         Restock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        store_id query    string true "Store ID"
// @Success      200      {object} response.WebResponse[[]model.RestockPredictionResponse]
// @Failure      400      {object} response.ApiErrorResponse
// @Failure      404      {object} response.ApiErrorResponse
// @Failure      500      {object} response.ApiErrorResponse
// @Router       /restock-predictions [get]
func (c *RestockController) List(ctx *fiber.Ctx) error {
	storeID := ctx.Query("store_id", "")

	resp, err := c.UseCase.List(ctx.UserContext(), storeID)
	if err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[[]model.RestockPredictionResponse]{
		Data:    resp,
		Message: "Berhasil menampilkan prediksi restock",
		Success: true,
	})
}
