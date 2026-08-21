ALTER TABLE transactions
    ADD COLUMN customer_id VARCHAR(36) NOT NULL AFTER store_id,
    ADD COLUMN customer_phone VARCHAR(20) NOT NULL DEFAULT '' AFTER customer_id,
    ADD COLUMN customer_name VARCHAR(100) NOT NULL DEFAULT '' AFTER customer_phone,
    ADD INDEX idx_transactions_customer_id (customer_id);
