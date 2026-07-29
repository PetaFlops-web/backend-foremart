package controller

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/report/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/report/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ReportController struct {
	Log     *logrus.Logger
	UseCase *usecase.ReportUseCase
}

func NewReportController(useCase *usecase.ReportUseCase, logger *logrus.Logger) *ReportController {
	return &ReportController{
		Log:     logger,
		UseCase: useCase,
	}
}

// Daily godoc
// @Summary      Menampilkan laporan harian
// @Description  Menghitung laporan harian toko secara on-the-fly berdasarkan transaksi dan stok terkini
// @Tags         Report
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        store_id query     string true  "Store ID"
// @Param        date     query     string false "Tanggal laporan (YYYY-MM-DD)"
// @Success      200      {object}  response.WebResponse[model.DailyReportResponse]
// @Failure      400      {object}  response.ApiErrorResponse
// @Failure      404      {object}  response.ApiErrorResponse
// @Failure      500      {object}  response.ApiErrorResponse
// @Router       /reports/daily [get]
func (c *ReportController) Daily(ctx *fiber.Ctx) error {
	request := &model.DailyReportRequest{
		StoreID: ctx.Query("store_id", ""),
		Date:    ctx.Query("date", ""),
	}

	resp, err := c.UseCase.Daily(ctx.UserContext(), request)
	if err != nil {
		c.Log.Warnf("Failed to generate daily report : %+v", err)
		return err
	}

	return ctx.JSON(response.WebResponse[*model.DailyReportResponse]{
		Data:    resp,
		Message: "Berhasil menampilkan laporan harian",
		Success: true,
	})
}
