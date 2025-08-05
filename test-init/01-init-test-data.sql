-- Sample initialization script for test database
-- This file will be executed when the container starts for the first time

-- Create a sample table for testing
CREATE TABLE IF NOT EXISTS test_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert some sample data
INSERT INTO test_users (name, email) VALUES 
    ('John Doe', 'john@example.com'),
    ('Jane Smith', 'jane@example.com'),
    ('Bob Johnson', 'bob@example.com')
ON CONFLICT (email) DO NOTHING;

-- Create another table for backup testing
CREATE TABLE IF NOT EXISTS test_orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES test_users(id),
    product_name VARCHAR(200) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample orders
INSERT INTO test_orders (user_id, product_name, price) VALUES 
    (1, 'Laptop', 999.99),
    (2, 'Mouse', 29.99),
    (3, 'Keyboard', 79.99),
    (1, 'Monitor', 299.99)
ON CONFLICT DO NOTHING;
