package model

type DailyReportRequest struct {
	StoreID string `json:"store_id" query:"store_id" validate:"required"`
	Date    string `json:"date" query:"date"`
}
