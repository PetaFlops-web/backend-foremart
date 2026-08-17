package converter

import (
	"time"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/model"
)

func RestockPredictionToResponse(prediction *entity.RestockPrediction) *model.RestockPredictionResponse {
	if prediction == nil {
		return nil
	}

	return &model.RestockPredictionResponse{
		ID:                    prediction.ID,
		StoreID:               prediction.StoreID,
		ProductID:             prediction.ProductID,
		ProductName:           prediction.ProductName,
		Unit:                  prediction.Unit,
		ForecastDate:          formatDate(prediction.ForecastDate),
		DailySales:            prediction.PredictedDailySales,
		CurrentStock:          prediction.CurrentStock,
		RecommendedRestockQty: prediction.RecommendedRestockQty,
		CreatedAt:             prediction.CreatedAt,
	}
}

func RestockPredictionsToResponses(predictions []entity.RestockPrediction) []model.RestockPredictionResponse {
	responses := make([]model.RestockPredictionResponse, len(predictions))
	for i := range predictions {
		responses[i] = *RestockPredictionToResponse(&predictions[i])
	}
	return responses
}

func formatDate(date *time.Time) string {
	if date == nil || date.IsZero() {
		return ""
	}
	return date.Format("2006-01-02")
}
