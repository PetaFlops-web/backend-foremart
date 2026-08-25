package model

type SurvivalPredictionRequest struct {
	StoreID    string `json:"store_id" validate:"required"`
	CustomerID string `json:"customer_id" validate:"required"`
	ProductID  string `json:"product_id" validate:"required"`
}
