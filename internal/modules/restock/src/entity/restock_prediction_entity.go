package entity

import "time"

type RestockPrediction struct {
	ID                    string     `gorm:"column:id;primaryKey;type:varchar(36)"`
	StoreID               string     `gorm:"column:store_id;type:varchar(36);not null;index:idx_restock_store_product_forecast_date,unique"`
	ProductID             string     `gorm:"column:product_id;type:varchar(36);not null;index:idx_restock_store_product_forecast_date,unique"`
	ProductName           string     `gorm:"column:product_name;type:varchar(255);not null"`
	Unit                  string     `gorm:"column:unit;type:varchar(50);not null;default:''"`
	ForecastDate          *time.Time `gorm:"column:forecast_date;type:date;not null;index:idx_restock_store_product_forecast_date,unique"`
	PredictedDailySales   int        `gorm:"column:predicted_daily_sales;not null;default:0"`
	CurrentStock          int        `gorm:"column:current_stock;not null;default:0"`
	RecommendedRestockQty int        `gorm:"column:recommended_restock_qty;not null;default:0"`
	HistoryDays           int        `gorm:"column:history_days;not null;default:30"`
	CreatedAt             int64      `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt             int64      `gorm:"column:updated_at;autoUpdateTime:milli"`
}


func (RestockPrediction) TableName() string {
	return "restock_predictions"
}
