ALTER TABLE transactions
    DROP INDEX idx_transactions_customer_id,
    DROP COLUMN customer_name,
    DROP COLUMN customer_phone,
    DROP COLUMN customer_id;
