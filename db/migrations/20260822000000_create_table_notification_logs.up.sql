CREATE TABLE IF NOT EXISTS notification_logs (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL,
    customer_id INT NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    channel VARCHAR(20) NOT NULL DEFAULT 'whatsapp',
    message TEXT NOT NULL,
    predicted_restock_date VARCHAR(10) NOT NULL DEFAULT '',
    rule_triggered VARCHAR(50) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'sent',
    period VARCHAR(10) NOT NULL,
    created_at BIGINT NOT NULL DEFAULT 0,
    INDEX idx_notification_logs_store (store_id),
    INDEX idx_notification_logs_customer (customer_id),
    INDEX idx_notification_logs_product (product_id),
    INDEX idx_notification_logs_period (period),
    UNIQUE KEY uq_notification_logs_dedup (customer_id, product_id, period)
);
