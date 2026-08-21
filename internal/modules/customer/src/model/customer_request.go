package model

type CreateCustomerRequest struct {
	StoreID string `json:"store_id" validate:"required"`
	Name    string `json:"name" validate:"omitempty,max=100"`
	Phone   string `json:"phone" validate:"required,min=8,max=20"`
}

type SearchCustomerRequest struct {
	StoreID string `json:"store_id" query:"store_id" validate:"required"`
	Search  string `json:"search"  query:"search"`
	Page    int    `json:"page"    query:"page"    validate:"required,min=1"`
	Size    int    `json:"size"    query:"size"    validate:"required,min=1,max=100"`
}
