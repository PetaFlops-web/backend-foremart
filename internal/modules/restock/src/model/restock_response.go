package model

type RestockPredictionResponse struct {
	ID                    string `json:"id"`
	StoreID               string `json:"store_id"`
	ProductID             string `json:"product_id"`
	ProductName           string `json:"product_name"`
	Unit                  string `json:"unit"`
	ForecastDate          string `json:"forecast_date"`
	DailySales            int    `json:"daily_sales"`
	CurrentStock          int    `json:"current_stock"`
	RecommendedRestockQty int    `json:"recommended_restock_qty"`
	CreatedAt             int64  `json:"created_at"`
}
type GenerateRestockPredictionResponse struct {
	GeneratedCount int                          `json:"generated_count"`
	SkippedCount   int                          `json:"skipped_count"`
	Items          []RestockPredictionResponse  `json:"items"`
	Skipped        []RestockSkippedItemResponse `json:"skipped"`
}

type RestockSkippedItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Reason      string `json:"reason"`
}
