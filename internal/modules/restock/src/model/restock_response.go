package model

type RestockPredictionResponse struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	CurrentStock int    `json:"current_stock"`
	RestockQty   int    `json:"restock_qty"`
	StockoutDate string `json:"stockout_date"`
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
