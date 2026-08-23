package entity

type NotificationLog struct {
	ID                   string `gorm:"column:id;primaryKey;type:varchar(36)"`
	StoreID              string `gorm:"column:store_id;type:varchar(36);not null;index:idx_notification_logs_store"`
	CustomerID           int    `gorm:"column:customer_id;type:int;not null;index:idx_notification_logs_customer;uniqueIndex:uq_notification_logs_dedup,priority:1"`
	ProductID            string `gorm:"column:product_id;type:varchar(36);not null;index:idx_notification_logs_product;uniqueIndex:uq_notification_logs_dedup,priority:2"`
	Channel              string `gorm:"column:channel;type:varchar(20);not null;default:'whatsapp'"`
	Message              string `gorm:"column:message;type:text;not null"`
	PredictedRestockDate string `gorm:"column:predicted_restock_date;type:varchar(10);not null;default:''"`
	RuleTriggered        string `gorm:"column:rule_triggered;type:varchar(50);not null;default:''"`
	Status               string `gorm:"column:status;type:varchar(20);not null;default:'sent'"`
	Period               string `gorm:"column:period;type:varchar(10);not null;index:idx_notification_logs_period;uniqueIndex:uq_notification_logs_dedup,priority:3"`
	CreatedAt            int64  `gorm:"column:created_at;autoCreateTime:milli"`
}

func (NotificationLog) TableName() string {
	return "notification_logs"
}
