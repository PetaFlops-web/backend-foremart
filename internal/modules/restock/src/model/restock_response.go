package model

type RestockPredictionResponse struct {
	ID                    string `json:"id"`
	StoreID               string `json:"store_id"`
	ProductID             string `json:"product_id"`
	ProductName           string `json:"product_name"`
	ForecastDate          string `json:"forecast_date"`
	PredictedDailySales   int    `json:"predicted_daily_sales"`
	CurrentStock          int    `json:"current_stock"`
	ForecastWindowDays    int    `json:"forecast_window_days"`
	RecommendedRestockQty int    `json:"recommended_restock_qty"`
	StockoutDate          string `json:"stockout_date"`
	HistoryDays           int    `json:"history_days"`
	CreatedAt             int64  `json:"created_at"`
	UpdatedAt             int64  `json:"updated_at"`
}

type GenerateRestockPredictionResponse struct {
	StoreID            string                       `json:"store_id"`
	ForecastDate       string                       `json:"forecast_date"`
	HistoryDays        int                          `json:"history_days"`
	ForecastWindowDays int                          `json:"forecast_window_days"`
	GeneratedCount     int                          `json:"generated_count"`
	SkippedCount       int                          `json:"skipped_count"`
	Items              []RestockPredictionResponse  `json:"items"`
	Skipped            []RestockSkippedItemResponse `json:"skipped"`
}

type RestockSkippedItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Reason      string `json:"reason"`
}
