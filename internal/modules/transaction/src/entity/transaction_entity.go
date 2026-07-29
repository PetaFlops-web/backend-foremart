package entity

import "time"

type Transaction struct {
	ID              string            `gorm:"column:id;primaryKey;type:varchar(36)"`
	StoreID         string            `gorm:"column:store_id;type:varchar(36);not null"`
	TransactionDate time.Time         `gorm:"column:transaction_date;type:date;not null"`
	Source          string            `gorm:"column:source;type:varchar(20);not null"`
	CreatedAt       int64             `gorm:"column:created_at;autoCreateTime:milli"`
	Items           []TransactionItem `gorm:"foreignKey:TransactionID"`
}

func (Transaction) TableName() string {
	return "transactions"
}