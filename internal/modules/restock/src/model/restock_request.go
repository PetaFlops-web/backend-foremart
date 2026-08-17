package model

type GenerateRestockPredictionRequest struct {
	StoreID     string `json:"store_id" validate:"required"`
	ProductID   string `json:"product_id,omitempty"`
	HistoryDays int    `json:"history_days,omitempty"`
	DaysAhead   int    `json:"days_ahead,omitempty"`
}
