package model

type SurvivalPredictionResponse struct {
	CustomerID             int     `json:"customer_id"`
	PurchaseNumber         int     `json:"purchase_number"`
	StockCode              string  `json:"stock_code"`
	PredictedRestockDate   string  `json:"predicted_restock_date"`
	PredDaysLeft           int     `json:"pred_days_left"`
	PredMedianSurvivalDays float64 `json:"pred_median_survival_days"`
	DaysSinceLastBuy       int     `json:"days_since_last_buy"`
	ProbBuyWithin7d        float64 `json:"prob_buy_within_7d"`
	ProbBuyWithin14d       float64 `json:"prob_buy_within_14d"`
	ProbBuyWithin30d       float64 `json:"prob_buy_within_30d"`
	PartialHazard          float64 `json:"partial_hazard"`
}
