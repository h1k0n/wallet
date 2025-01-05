package main

import (
	"database/sql"
	"time"
)

type Wallet struct {
	ID      int64   `json:"id"`
	Balance float64 `json:"balance"`
	UserID  string  `json:"user_id"`
}

type Transaction struct {
	ID        int64     `json:"id"`
	WalletID  int64     `json:"wallet_id"`
	OpType    string    `json:"op_type"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type IWallet interface {
	UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error
	ExecTransfer(db *sql.DB, fromWalletID, toWalletID int64, amount float64) error
	GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error)
	GetTransactionsByWalletID(db *sql.DB, walletID int64, limit, offset int) ([]Transaction, error)
}
