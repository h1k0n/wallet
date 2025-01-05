package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetTransactionsByWalletID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	walletID := int64(1)
	limit := 10
	offset := 0

	rows := sqlmock.NewRows([]string{"id", "wallet_id", "op_type", "amount", "created_at"}).
		AddRow(1, walletID, "deposit", 100.0, time.Now()).
		AddRow(2, walletID, "withdraw", 50.0, time.Now())

	mock.ExpectQuery("SELECT id, wallet_id, op_type, amount, created_at FROM transactions WHERE wallet_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
		WithArgs(walletID, limit, offset).
		WillReturnRows(rows)
	wa := &WalletAccess{}
	transactions, err := wa.GetTransactionsByWalletID(db, walletID, limit, offset)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(transactions))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestGetWalletInfoById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	walletID := int64(1)
	expectedWallet := &Wallet{
		ID:      walletID,
		Balance: 100.0,
		UserID:  "1",
	}

	rows := sqlmock.NewRows([]string{"id", "balance", "user_id"}).
		AddRow(expectedWallet.ID, expectedWallet.Balance, expectedWallet.UserID)

	mock.ExpectQuery("SELECT id, balance, user_id FROM wallet WHERE id = \\$1").
		WithArgs(walletID).
		WillReturnRows(rows)
	wa := &WalletAccess{}
	wallet, err := wa.GetWalletInfoById(db, walletID)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if wallet.ID != expectedWallet.ID || wallet.Balance != expectedWallet.Balance || wallet.UserID != expectedWallet.UserID {
		t.Errorf("expected wallet %+v, got %+v", expectedWallet, wallet)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetWalletInfoById_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	walletID := int64(1)

	mock.ExpectQuery("SELECT id, balance, user_id FROM wallet WHERE id = \\$1").
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)
	wa := &WalletAccess{}
	_, err = wa.GetWalletInfoById(db, walletID)
	if err == nil || err.Error() != "wallet not found" {
		t.Errorf("expected 'wallet not found' error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestUpdateBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	walletID := int64(1)
	opType := "deposit"
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, walletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(walletID, opType, amount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	wa := &WalletAccess{}
	err = wa.UpdateBalance(db, walletID, opType, amount)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateBalance_TransactionFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	walletID := int64(1)
	opType := "deposit"
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, walletID).
		WillReturnError(fmt.Errorf("update failed"))

	mock.ExpectRollback()
	wa := &WalletAccess{}
	err = wa.UpdateBalance(db, walletID, opType, amount)
	if err == nil || err.Error() != "update failed" {
		t.Errorf("expected 'update failed' error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateBalance_InsertTransactionFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	walletID := int64(1)
	opType := "deposit"
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, walletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(walletID, opType, amount, sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert transaction failed"))

	mock.ExpectRollback()
	wa := &WalletAccess{}
	err = wa.UpdateBalance(db, walletID, opType, amount)
	if err == nil || err.Error() != "insert transaction failed" {
		t.Errorf("expected 'insert transaction failed' error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestPerformTransfer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, toWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(fromWalletID, "transfer", -amount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(toWalletID, "transfer", amount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = performTransfer(tx, fromWalletID, toWalletID, amount)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPerformTransfer_DeductFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnError(fmt.Errorf("deduct failed"))

	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = performTransfer(tx, fromWalletID, toWalletID, amount)
	if err == nil || err.Error() != "failed to deduct from sender's balance: deduct failed" {
		t.Errorf("expected 'failed to deduct from sender's balance: deduct failed' error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPerformTransfer_AddFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, toWalletID).
		WillReturnError(fmt.Errorf("add failed"))

	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = performTransfer(tx, fromWalletID, toWalletID, amount)
	if err == nil || err.Error() != "failed to add to receiver's balance: add failed" {
		t.Errorf("expected 'failed to add to receiver's balance: add failed' error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPerformTransfer_InsertSenderTransactionFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, toWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(fromWalletID, "transfer", -amount, sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert sender transaction failed"))

	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = performTransfer(tx, fromWalletID, toWalletID, amount)
	if err == nil || err.Error() != "failed to insert sender's transaction: insert sender transaction failed" {
		t.Errorf("expected 'failed to insert sender's transaction: insert sender transaction failed' error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPerformTransfer_InsertReceiverTransactionFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, toWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(fromWalletID, "transfer", -amount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(toWalletID, "transfer", amount, sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert receiver transaction failed"))

	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = performTransfer(tx, fromWalletID, toWalletID, amount)
	if err == nil || err.Error() != "failed to insert receiver's transaction: insert receiver transaction failed" {
		t.Errorf("expected 'failed to insert receiver's transaction: insert receiver transaction failed' error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestLockwalletForTransfer_LockFromWalletFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)

	mock.ExpectBegin()

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(fromWalletID).
		WillReturnError(fmt.Errorf("lock from wallet failed"))

	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = lockwalletForTransfer(tx, fromWalletID, toWalletID)
	if err == nil || err.Error() != "lock from wallet failed" {
		t.Errorf("expected 'lock from wallet failed' error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestLockwalletForTransfer_LockToWalletFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)

	mock.ExpectBegin()

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(toWalletID).
		WillReturnError(fmt.Errorf("lock to wallet failed"))

	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = lockwalletForTransfer(tx, fromWalletID, toWalletID)
	if err == nil || err.Error() != "lock to wallet failed" {
		t.Errorf("expected 'lock to wallet failed' error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestExecTransfer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(toWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance \\+ \\$1 WHERE id = \\$2").
		WithArgs(amount, toWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(fromWalletID, "transfer", -amount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO transactions \\(wallet_id, op_type, amount, created_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
		WithArgs(toWalletID, "transfer", amount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	wa := &WalletAccess{}
	err = wa.ExecTransfer(db, fromWalletID, toWalletID, amount)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestExecTransfer_LockFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(fromWalletID).
		WillReturnError(fmt.Errorf("lock from wallet failed"))

	mock.ExpectRollback()

	wa := &WalletAccess{}
	err = wa.ExecTransfer(db, fromWalletID, toWalletID, amount)
	if err == nil || err.Error() != "lock from wallet failed" {
		t.Errorf("expected 'lock from wallet failed' error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestExecTransfer_TransferFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fromWalletID := int64(1)
	toWalletID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(fromWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("SELECT 1 FROM wallet WHERE id = \\$1 FOR UPDATE").
		WithArgs(toWalletID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE wallet SET balance = balance - \\$1 WHERE id = \\$2").
		WithArgs(amount, fromWalletID).
		WillReturnError(fmt.Errorf("deduct failed"))

	mock.ExpectRollback()

	wa := &WalletAccess{}
	err = wa.ExecTransfer(db, fromWalletID, toWalletID, amount)
	if err == nil || err.Error() != "failed to deduct from sender's balance: deduct failed" {
		t.Errorf("expected 'failed to deduct from sender's balance: deduct failed' error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
