package model

// TransactionPreviewItemResponse is a single item in the preview, merging ML + DB data.
type TransactionPreviewItemResponse struct {
	// From ML
	RawText       string  `json:"raw_text"`
	DetectedQty   float64 `json:"detected_qty"`
	DetectedPrice int64   `json:"detected_price"`

	// From Database (matched product)
	ProductId    string `json:"product_id,omitempty"`
	ProductName  string `json:"product_name,omitempty"`
	SellingPrice int64  `json:"selling_price,omitempty"`
	CostPrice    int64  `json:"cost_price,omitempty"`
	Stock        int    `json:"stock,omitempty"`

	// Match status
	IsMatched bool `json:"is_matched"`
}

// TransactionPreviewResponse wraps the full preview returned to the Frontend.
type TransactionPreviewResponse struct {
	RawText string                           `json:"raw_text"`
	Items   []TransactionPreviewItemResponse `json:"items"`
}

// TransactionResponse is returned after a transaction is confirmed/saved.
type TransactionResponse struct {
	ID              string                        `json:"id"`
	StoreID         string                        `json:"store_id"`
	TransactionDate string                        `json:"transaction_date"`
	Source          string                        `json:"source"`
	CreatedAt       int64                         `json:"created_at"`
	Items           []TransactionItemResponse     `json:"items,omitempty"`
}