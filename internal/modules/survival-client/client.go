package survival_client

import (
	"context"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/survival/src/model"
)

// Client is the public contract for triggering survival predictions.
type Client interface {
	Predict(ctx context.Context, req *model.SurvivalPredictionRequest) (*model.SurvivalPredictionResponse, error)
}
