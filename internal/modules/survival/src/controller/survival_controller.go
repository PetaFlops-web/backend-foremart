package controller

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SurvivalController struct {
	Log     *logrus.Logger
	UseCase *usecase.SurvivalUseCase
}

func NewSurvivalController(useCase *usecase.SurvivalUseCase, logger *logrus.Logger) *SurvivalController {
	return &SurvivalController{
		Log:     logger,
		UseCase: useCase,
	}
}

// Predict godoc
// @Summary      Prediksi pembelian ulang
// @Description  Menghitung fitur survival dari riwayat transaksi lalu memanggil ML untuk memprediksi kapan customer membeli ulang produk
// @Tags         Survival
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body     model.SurvivalPredictionRequest true "Parameter prediksi survival"
// @Success      200  {object} response.WebResponse[model.SurvivalPredictionResponse]
// @Failure      400  {object} response.ApiErrorResponse
// @Failure      403  {object} response.ApiErrorResponse
// @Failure      404  {object} response.ApiErrorResponse
// @Failure      500  {object} response.ApiErrorResponse
// @Router       /predict-survival [post]
func (c *SurvivalController) Predict(ctx *fiber.Ctx) error {
	request := new(model.SurvivalPredictionRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	resp, err := c.UseCase.Predict(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	return ctx.JSON(response.WebResponse[*model.SurvivalPredictionResponse]{
		Data:    resp,
		Message: "Berhasil membuat prediksi pembelian ulang",
		Success: true,
	})
}
