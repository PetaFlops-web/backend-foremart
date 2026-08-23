package controller

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type NotificationController struct {
	Log     *logrus.Logger
	UseCase *usecase.NotificationUseCase
}

func NewNotificationController(useCase *usecase.NotificationUseCase, logger *logrus.Logger) *NotificationController {
	return &NotificationController{
		Log:     logger,
		UseCase: useCase,
	}
}

// List godoc
// @Summary      Daftar log notifikasi pengingat pembelian ulang
// @Description  Mengembalikan log notifikasi yang sudah dikirim untuk suatu toko
// @Tags         Notification
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        store_id query string true "ID toko"
// @Success      200  {object} response.WebResponse[[]model.NotificationResponse]
// @Failure      400  {object} response.ApiErrorResponse
// @Router       /notifications [get]
func (c *NotificationController) List(ctx *fiber.Ctx) error {
	storeID := ctx.Query("store_id")
	if storeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "store_id wajib diisi")
	}

	logs, err := c.UseCase.List(ctx.UserContext(), storeID)
	if err != nil {
		c.Log.Errorf("Failed to list notification logs: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil log notifikasi")
	}

	return ctx.JSON(response.WebResponse[[]model.NotificationResponse]{
		Data:    logs,
		Message: "Berhasil mengambil log notifikasi",
		Success: true,
	})
}

// Send godoc
// @Summary      Jalankan pengiriman notifikasi pembelian ulang (manual)
// @Description  Memicu proses pengiriman notifikasi WhatsApp secara manual untuk seluruh toko (mirror cron job)
// @Tags         Notification
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} response.WebResponse[int]
// @Failure      500  {object} response.ApiErrorResponse
// @Router       /notifications/_send [post]
func (c *NotificationController) Send(ctx *fiber.Ctx) error {
	sent, err := c.UseCase.RunReminder(ctx.UserContext())
	if err != nil {
		c.Log.Errorf("Failed to run reorder reminder: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menjalankan pengiriman notifikasi")
	}

	return ctx.JSON(response.WebResponse[int]{
		Data:    sent,
		Message: "Berhasil menjalankan pengiriman notifikasi",
		Success: true,
	})
}
