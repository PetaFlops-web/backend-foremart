package model

// TransactionItemResponse is a single item in a saved transaction.
type TransactionItemResponse struct {
	ID                   string `json:"id"`
	ProductID            string `json:"product_id"`
	ProductNameSnapshot  string `json:"product_name_snapshot"`
	Qty                  int    `json:"qty"`
	CostPriceSnapshot    int64  `json:"cost_price_snapshot"`
	SellingPriceSnapshot int64  `json:"selling_price_snapshot"`
}