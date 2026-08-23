package converter

import (
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/model"
)

func NotificationToResponse(e *entity.NotificationLog) *model.NotificationResponse {
	return &model.NotificationResponse{
		ID:                   e.ID,
		StoreID:              e.StoreID,
		CustomerID:           e.CustomerID,
		ProductID:            e.ProductID,
		Channel:              e.Channel,
		Message:              e.Message,
		PredictedRestockDate: e.PredictedRestockDate,
		RuleTriggered:        e.RuleTriggered,
		Status:               e.Status,
		Period:               e.Period,
		CreatedAt:            e.CreatedAt,
	}
}
