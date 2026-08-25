package entity

import "time"

type Transaction struct {
	ID              string            `gorm:"column:id;primaryKey;type:varchar(36)"`
	StoreID         string            `gorm:"column:store_id;type:varchar(36);not null"`
	CustomerID      string            `gorm:"column:customer_id;type:varchar(36);index:idx_transactions_customer_id"`
	CustomerPhone   string            `gorm:"column:customer_phone;type:varchar(20);not null;default:''"`
	CustomerName    string            `gorm:"column:customer_name;type:varchar(100);not null;default:''"`
	TransactionDate time.Time         `gorm:"column:transaction_date;type:date;not null"`
	Source          string            `gorm:"column:source;type:varchar(20);not null"`
	CreatedAt       int64             `gorm:"column:created_at;autoCreateTime:milli"`
	Items           []TransactionItem `gorm:"foreignKey:TransactionID"`
}

func (Transaction) TableName() string {
	return "transactions"
}