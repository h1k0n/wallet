-- Create the wallet table to store user wallet information
CREATE TABLE IF NOT EXISTS wallet (
    id SERIAL PRIMARY KEY, -- Unique identifier for each wallet
    balance DECIMAL(10, 2) DEFAULT 0.00, -- Wallet balance with a default value of 0.00
    user_id VARCHAR(255) NOT NULL -- User ID associated with the wallet
);

COMMENT ON COLUMN wallet.id IS 'Unique identifier for each wallet';
COMMENT ON COLUMN wallet.balance IS 'Wallet balance with a default value of 0.00';
COMMENT ON COLUMN wallet.user_id IS 'User ID associated with the wallet';

-- Create the transactions table to store transaction details
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY, -- Unique identifier for each transaction
    wallet_id INT, -- Foreign key referencing the wallet table
    op_type VARCHAR(20) CHECK (op_type IN ('deposit', 'withdraw', 'transfer')) NOT NULL, -- Type of transaction: 'deposit', 'withdraw', or 'transfer'
    amount DECIMAL(10, 2) NOT NULL, -- Amount involved in the transaction
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of the transaction, defaults to the current time
    FOREIGN KEY (wallet_id) REFERENCES wallet(id) -- Foreign key constraint linking to the wallet table
);

COMMENT ON COLUMN transactions.id IS 'Unique identifier for each transaction';
COMMENT ON COLUMN transactions.wallet_id IS 'Foreign key referencing the wallet table';
COMMENT ON COLUMN transactions.op_type IS 'Type of transaction: deposit, withdraw, or transfer';
COMMENT ON COLUMN transactions.amount IS 'Amount involved in the transaction';
COMMENT ON COLUMN transactions.created_at IS 'Timestamp of the transaction, defaults to the current time';