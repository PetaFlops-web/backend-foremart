package model

type SurvivalPredictionRequest struct {
	StoreID    string `json:"store_id" validate:"required"`
	CustomerID int    `json:"customer_id" validate:"required,min=1"`
	ProductID  string `json:"product_id" validate:"required"`
}
