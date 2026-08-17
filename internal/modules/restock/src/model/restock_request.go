package model

type GenerateRestockPredictionRequest struct {
	StoreID            string `json:"store_id" validate:"required"`
	ProductID          string `json:"product_id,omitempty"`
	ForecastDate       string `json:"forecast_date,omitempty"`
	HistoryDays        int    `json:"history_days,omitempty"`
	ForecastWindowDays int    `json:"forecast_window_days,omitempty"`
}
