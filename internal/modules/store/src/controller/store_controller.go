package controller

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/model"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store/src/usecase"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/middleware"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type StoreController struct {
	UseCase *usecase.StoreUseCase
	Log     *logrus.Logger
}

func NewStoreController(useCase *usecase.StoreUseCase, logger *logrus.Logger) *StoreController {
	return &StoreController{
		UseCase: useCase,
		Log:     logger,
	}
}

// Create godoc
// @Summary      Create a new store
// @Description  Creates a new store for the currently authenticated user
// @Tags         Store
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body model.StoreRequest true "Store details"
// @Success      200  {object}  response.WebResponse[model.StoreResponse]
// @Failure      400  {object}  response.ApiErrorResponse
// @Failure      401  {object}  response.ApiErrorResponse
// @Failure      409  {object}  response.ApiErrorResponse
// @Router       /stores [post]
func (c *StoreController) Create(ctx *fiber.Ctx) error {
	auth := middleware.GetUser(ctx)

	request := new(model.StoreRequest)
	if err := ctx.BodyParser(&request); err != nil {
		c.Log.Warnf("Failed to parse request body : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Format data request tidak valid")
	}

	request.UserID = auth.ID

	resp, err := c.UseCase.Create(ctx.UserContext(), request)
	if err != nil {
		c.Log.WithError(err).Error("error creating store")
		return err
	}

	c.Log.Infof("Store created successfully with ID: %s, User: %s", resp.ID, resp.UserID)

	return ctx.JSON(response.WebResponse[*model.StoreResponse]{
		Data:    resp,
		Message: "Berhasil menambahkan store",
		Success: true,
	})
}

// GetByUserID godoc
// @Summary      Get my store
// @Description  Returns the store of the currently authenticated user
// @Tags         Store
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.WebResponse[model.StoreResponse]
// @Failure      401  {object}  response.ApiErrorResponse
// @Failure      404  {object}  response.ApiErrorResponse
// @Router       /stores/me [get]
func (c *StoreController) GetByUserID(ctx *fiber.Ctx) error {
	auth := middleware.GetUser(ctx)
	resp, err := c.UseCase.GetByUserID(ctx.UserContext(), auth.ID)
	if err != nil {
		c.Log.WithError(err).Error("error getting store by user ID")
		return err
	}
	return ctx.JSON(response.WebResponse[*model.StoreResponse]{
		Data:    resp,
		Message: "Berhasil mendapatkan store",
		Success: true,
	})
}

// GetMyStore godoc
// @Summary      Get
// @Description  Returns the store of the currently authenticated user
// @Tags         Store
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.WebResponse[model.StoreResponse]
// @Failure      401  {object}  response.ApiErrorResponse
// @Failure      404  {object}  response.ApiErrorResponse
// @Router       /stores [get]
func (c *StoreController) Get(ctx *fiber.Ctx) error {
	auth := middleware.GetUser(ctx)
	id := ctx.Params("id")

	resp, err := c.UseCase.Get(ctx.UserContext(), auth.ID, id)
	if err != nil {
		c.Log.WithError(err).Error("error getting my store")
		return err
	}

	return ctx.JSON(response.WebResponse[*model.StoreResponse]{
		Data:    resp,
		Message: "Berhasil mendapatkan store",
		Success: true,
	})
}

// UpdateMyStore godoc
// @Summary      Update
// @Description  Updates the store of the currently authenticated user
// @Tags         Store
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body model.StoreRequest true "Store update details"
// @Success      200  {object}  response.WebResponse[model.StoreResponse]
// @Failure      400  {object}  response.ApiErrorResponse
// @Failure      401  {object}  response.ApiErrorResponse
// @Failure      404  {object}  response.ApiErrorResponse
// @Router       /stores [put]
func (c *StoreController) Update(ctx *fiber.Ctx) error {
	auth := middleware.GetUser(ctx)

	request := new(model.StoreRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithError(err).Error("error parsing request body")
		return fiber.ErrBadRequest
	}
	request.ID = ctx.Params("id")
	request.UserID = auth.ID

	resp, err := c.UseCase.Update(ctx.UserContext(), request)
	if err != nil {
		c.Log.WithError(err).Error("error updating store")
		return err
	}

	return ctx.JSON(response.WebResponse[*model.StoreResponse]{
		Data:    resp,
		Message: "Berhasil mengubah data store",
		Success: true,
	})
}
