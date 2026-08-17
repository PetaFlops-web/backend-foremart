package repository

import (
	"context"
	"time"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/restock/src/entity"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RestockRepository struct {
	repository.Repository[entity.RestockPrediction]
	Log *logrus.Logger
	db  *gorm.DB
}

func NewRestockRepository(log *logrus.Logger, db *gorm.DB) *RestockRepository {
	return &RestockRepository{Log: log, db: db}
}

func (r *RestockRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *RestockRepository) UpsertByStoreProductDate(ctx context.Context, prediction *entity.RestockPrediction) (*entity.RestockPrediction, error) {
	db := r.dbWithContext(ctx)
	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "store_id"},
			{Name: "product_id"},
			{Name: "forecast_date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"product_name",
			"unit",
			"predicted_daily_sales",
			"current_stock",
			"recommended_restock_qty",
			"history_days",
			"updated_at",
		}),
	}).Create(prediction).Error
	if err != nil {
		return nil, err
	}

	saved := new(entity.RestockPrediction)
	if err := r.FindByStoreProductDate(ctx, saved, prediction.StoreID, prediction.ProductID, prediction.ForecastDate); err != nil {
		return nil, err
	}

	return saved, nil
}

func (r *RestockRepository) FindByStoreProductDate(ctx context.Context, prediction *entity.RestockPrediction, storeID string, productID string, forecastDate *time.Time) error {
	return r.dbWithContext(ctx).Where("store_id = ? AND product_id = ? AND forecast_date = ?", storeID, productID, forecastDate.Format("2006-01-02")).Take(prediction).Error
}

func (r *RestockRepository) ListByStoreID(ctx context.Context, storeID string) ([]entity.RestockPrediction, error) {
	var predictions []entity.RestockPrediction
	err := r.dbWithContext(ctx).Where("store_id = ?", storeID).
		Order("created_at DESC").
		Order("product_name ASC").
		Find(&predictions).Error
	return predictions, err
}
