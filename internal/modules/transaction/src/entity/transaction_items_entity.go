package entity

type TransactionItem struct {
	ID                   string `gorm:"column:id;primaryKey;type:varchar(36)"`
	TransactionID        string `gorm:"column:transaction_id;type:varchar(36);not null"`
	ProductID            string `gorm:"column:product_id;type:varchar(36);not null"`
	ProductNameSnapshot  string `gorm:"column:product_name_snapshot;type:varchar(255);not null"`
	Qty                  int    `gorm:"column:qty;type:int;not null"`
	CostPriceSnapshot    int64  `gorm:"column:cost_price_snapshot;type:bigint;not null"`
	SellingPriceSnapshot int64  `gorm:"column:selling_price_snapshot;type:bigint;not null"`
}

func (TransactionItem) TableName() string {
	return "transaction_items"
}