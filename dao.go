package main

import (
	"database/sql"
	"fmt"
	"time"

	"log"
)

type WalletAccess struct{}

func (wa *WalletAccess) UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// 更新钱包余额
	_, err = tx.Exec("UPDATE wallet SET balance = balance + $1 WHERE id = $2", amount, walletID)
	if err != nil {
		return err
	}

	// 插入交易记录
	_, err = tx.Exec("INSERT INTO transactions (wallet_id, op_type, amount, created_at) VALUES ($1, $2, $3, $4)", walletID, opType, amount, time.Now())
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

// 锁定发起账户和接收账户
func lockwalletForTransfer(tx *sql.Tx, fromWalletID int64, toWalletID int64) error {
	_, err := tx.Exec("SELECT 1 FROM wallet WHERE id = $1 FOR UPDATE", fromWalletID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("SELECT 1 FROM wallet WHERE id = $1 FOR UPDATE", toWalletID)
	return err
}

// 执行转账操作
func performTransfer(tx *sql.Tx, fromWalletID int64, toWalletID int64, amount float64) error {
	// 扣除发起账户的余额
	_, err := tx.Exec("UPDATE wallet SET balance = balance - $1 WHERE id = $2", amount, fromWalletID)
	if err != nil {
		return fmt.Errorf("failed to deduct from sender's balance: %v", err)
	}

	// 增加接收账户的余额
	_, err = tx.Exec("UPDATE wallet SET balance = balance + $1 WHERE id = $2", amount, toWalletID)
	if err != nil {
		return fmt.Errorf("failed to add to receiver's balance: %v", err)
	}

	// 插入发起账户的交易记录
	_, err = tx.Exec("INSERT INTO transactions (wallet_id, op_type, amount, created_at) VALUES ($1, $2, $3, $4)", fromWalletID, "transfer", -amount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert sender's transaction: %v", err)
	}

	// 插入接收账户的交易记录
	_, err = tx.Exec("INSERT INTO transactions (wallet_id, op_type, amount, created_at) VALUES ($1, $2, $3, $4)", toWalletID, "transfer", amount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert receiver's transaction: %v", err)
	}

	return nil
}

func (wa *WalletAccess) ExecTransfer(db *sql.DB, fromId, toId int64, amount float64) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// 锁定发起钱包和接收钱包
	err = lockwalletForTransfer(tx, fromId, toId)
	if err != nil {
		return err
	}

	// 执行转账操作
	err = performTransfer(tx, fromId, toId, amount)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

// 根据钱包id获取钱包信息
func (wa *WalletAccess) GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error) {
	var wallet Wallet
	err := db.QueryRow("SELECT id, balance, user_id FROM wallet WHERE id = $1", walletID).
		Scan(&wallet.ID, &wallet.Balance, &wallet.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}
	return &wallet, nil
}

// 根据钱包 ID 获取交易记录
func (wa *WalletAccess) GetTransactionsByWalletID(db *sql.DB, walletID int64, limit int, offset int) ([]Transaction, error) {
	rows, err := db.Query(`
		SELECT id, wallet_id, op_type, amount, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.WalletID, &tx.OpType, &tx.Amount, &tx.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
