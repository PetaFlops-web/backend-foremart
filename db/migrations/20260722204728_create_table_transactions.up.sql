CREATE TABLE transactions (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL,
    transaction_date DATE NOT NULL,
    source VARCHAR(20) NOT NULL,
    created_at BIGINT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE transaction_items (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    product_name_snapshot VARCHAR(255) NOT NULL,
    qty INT NOT NULL,
    cost_price_snapshot BIGINT NOT NULL,
    selling_price_snapshot BIGINT NOT NULL,
    CONSTRAINT fk_transaction_items_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
