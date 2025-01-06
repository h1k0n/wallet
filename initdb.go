package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

//var db *sql.DB

func (a *App) initDB(user, password, dbname, host string) {
	var err error
	// PostgreSQL 连接字符串
	dsn := fmt.Sprintf("user=%v password=%v dbname=%v host=%v sslmode=disable", user, password, dbname, host)
	a.DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err := a.DB.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	a.ensureTableExists()
	a.DB.SetMaxOpenConns(500)
	a.DB.SetMaxIdleConns(500)
	a.Rp = &WalletAccess{}
}

func (a *App) ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

const tableCreationQuery = `-- Create the wallet table to store user wallet information
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

insert into wallet values(1,0,'user1') ON CONFLICT (id) DO NOTHING;
insert into wallet values(2,0,'user2') ON CONFLICT (id) DO NOTHING;
`
