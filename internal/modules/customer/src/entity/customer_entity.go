package entity

import "gorm.io/gorm"

type Customer struct {
	ID               string `gorm:"column:id;primaryKey"`
	NumericID        int64  `gorm:"column:numeric_id;uniqueIndex:uni_customers_numeric_id"`
	Name             string `gorm:"column:name;type:varchar(100);not null;default:''"`
	Phone            string `gorm:"column:phone;type:varchar(20);not null;uniqueIndex:idx_customers_store_phone"`
	CreatedByStoreID string `gorm:"column:created_by_store_id;type:varchar(36);not null;index:idx_customers_created_by_store"`
	CreatedAt        int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        int64  `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (Customer) TableName() string {
	return "customers"
}

// BeforeCreate assigns a monotonically increasing numeric_id (surrogate key)
// used to satisfy the ML survival service, which expects an integer customer_id.
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.NumericID != 0 {
		return nil
	}
	var max int64
	if err := tx.Model(&Customer{}).Select("COALESCE(MAX(numeric_id), 0)").Scan(&max).Error; err != nil {
		return err
	}
	c.NumericID = max + 1
	return nil
}
