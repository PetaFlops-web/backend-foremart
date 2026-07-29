package model

type StoreSummaryResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	StoreName string `json:"store_name"`
}

type BestSellingProductResponse struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	QtySold      int    `json:"qty_sold"`
	CurrentStock int    `json:"current_stock"`
	Unit         string `json:"unit"`
}

type StockSummaryResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Stock       int    `json:"stock"`
	Unit        string `json:"unit"`
}

type DailyReportResponse struct {
	Store           StoreSummaryResponse         `json:"store"`
	Date            string                       `json:"date"`
	TotalOmset      int64                        `json:"total_omset"`
	TotalUntung     int64                        `json:"total_untung"`
	JumlahTransaksi int                          `json:"jumlah_transaksi"`
	ProdukTerlaris  []BestSellingProductResponse `json:"produk_terlaris"`
	SisaStok        []StockSummaryResponse       `json:"sisa_stok"`
}
