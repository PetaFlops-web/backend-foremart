package model

type NotificationResponse struct {
	ID                   string `json:"id"`
	StoreID              string `json:"store_id"`
	CustomerID           string `json:"customer_id"`
	ProductID            string `json:"product_id"`
	Channel              string `json:"channel"`
	Message              string `json:"message"`
	PredictedRestockDate string `json:"predicted_restock_date"`
	RuleTriggered        string `json:"rule_triggered"`
	Status               string `json:"status"`
	Period               string `json:"period"`
	CreatedAt            int64  `json:"created_at"`
}
