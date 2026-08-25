package model

// SeedMockRequest is used to generate backdated mock transactions that
// exercise the ML pipeline (restock prediction and survival/notification).
type SeedMockRequest struct {
	StoreID string `json:"store_id" validate:"required"`
	Mode    string `json:"mode" validate:"required,oneof=restock survival"`
}

// SeedMockResponse reports what the mock generator created.
type SeedMockResponse struct {
	Mode                string `json:"mode"`
	TransactionsCreated int    `json:"transactions_created"`
	ProductsAffected    int    `json:"products_affected"`
	CustomersAffected   int    `json:"customers_affected"`
}
