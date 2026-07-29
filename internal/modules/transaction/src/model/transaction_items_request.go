package model

// TransactionItemRequest is a single item in a confirmed transaction.
type TransactionItemRequest struct {
	ProductId         string `json:"product_id" validate:"required"`
	Qty               int    `json:"qty" validate:"required,min=1"`
	SellingPriceFinal int64  `json:"selling_price_final" validate:"required,min=0"`
}