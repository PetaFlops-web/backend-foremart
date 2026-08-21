package entity

type Customer struct {
	ID               string `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name             string `gorm:"column:name;type:varchar(100);not null;default:''"`
	Phone            string `gorm:"column:phone;type:varchar(20);not null;uniqueIndex:idx_customers_phone"`
	CreatedByStoreID string `gorm:"column:created_by_store_id;type:varchar(36);not null;index:idx_customers_created_by_store"`
	CreatedAt        int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        int64  `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (Customer) TableName() string {
	return "customers"
}
