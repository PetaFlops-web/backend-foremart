CREATE TABLE restock_predictions (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    prediction_date DATE NOT NULL,
    predicted_sales INT NOT NULL DEFAULT 0,
    current_stock INT NOT NULL DEFAULT 0,
    forecast_window_days INT NOT NULL DEFAULT 7,
    recommended_qty INT NOT NULL DEFAULT 0,
    predicted_restock_date DATE NULL,
    history_days INT NOT NULL DEFAULT 30,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    UNIQUE INDEX idx_restock_store_product_date (store_id, product_id, prediction_date),
    INDEX idx_restock_store_created (store_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
