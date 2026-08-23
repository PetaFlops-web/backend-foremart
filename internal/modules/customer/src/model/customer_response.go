package model

type CustomerResponse struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	CreatedByStoreID string `json:"created_by_store_id"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}
