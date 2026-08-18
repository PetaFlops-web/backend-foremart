package model

import "time"

type GenerateRestockPredictionRequest struct {
	StoreID      string     `json:"store_id" validate:"required"`
	ProductID    string     `json:"product_id,omitempty"`
	ForecastDate *time.Time `json:"forecast_date,omitempty"`
	HistoryDays  int        `json:"history_days,omitempty"`
}
