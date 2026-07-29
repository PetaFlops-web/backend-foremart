package model

// ExtractVoiceRequest is used by the Controller to pass audio data to the UseCase.
type ExtractVoiceRequest struct {
	StoreId   string `json:"store_id" validate:"required"`
	AudioData []byte `json:"-"`
	Filename  string `json:"-"`
}

// CreateTransactionRequest is used when the user confirms a transaction after preview.
type CreateTransactionRequest struct {
	StoreId string                   `json:"store_id" validate:"required"`
	Source  string                   `json:"source" validate:"required"`
	Items   []TransactionItemRequest `json:"items" validate:"required,dive"`
}

// SearchTransactionRequest is used for listing/paginating transactions.
type SearchTransactionRequest struct {
	StoreId string `json:"store_id" query:"store_id" validate:"required"`
	Page    int    `json:"page" query:"page" validate:"required,min=1"`
	Size    int    `json:"size" query:"size" validate:"required,min=1,max=100"`
}