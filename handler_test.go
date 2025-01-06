package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockWalletRepo struct{}

func (m *MockWalletRepo) UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error {
	if walletID == 1 {
		return nil
	}
	return errors.New("update balance failed")
}

func (m *MockWalletRepo) ExecTransfer(db *sql.DB, fromWalletID, toWalletID int64, amount float64) error {
	return nil
}

func (m *MockWalletRepo) GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error) {
	if walletID == 1 {
		return &Wallet{ID: 1, Balance: 100.0}, nil
	}
	return nil, errors.New("wallet not found")
}

func TestDepositWithdrawHandler(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletRepo{}}
	router.PUT("/api/balance/:id", a.depositWithdrawHandler)

	// Test cases
	tests := []struct {
		name           string
		id             string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Deposit Success",
			id:   "1",
			requestBody: map[string]interface{}{
				"op_type": "deposit",
				"amount":  100.0,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "deposit successful"},
		},
		{
			name: "Withdraw Success",
			id:   "1",
			requestBody: map[string]interface{}{
				"op_type": "withdraw",
				"amount":  50.0,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "withdraw successful"},
		},
		{
			name: "Invalid Operation Type",
			id:   "1",
			requestBody: map[string]interface{}{
				"op_type": "invalid",
				"amount":  50.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "invalid operation type"},
		},
		{
			name: "Wallet Not Found",
			id:   "999",
			requestBody: map[string]interface{}{
				"op_type": "deposit",
				"amount":  100.0,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "wallet not found"},
		},
		{
			name: "Not Enough Balance",
			id:   "1",
			requestBody: map[string]interface{}{
				"op_type": "withdraw",
				"amount":  500.0,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "not enough"},
		},
		{
			name: "invalid id",
			id:   "1k",
			requestBody: map[string]interface{}{
				"op_type": "withdraw",
				"amount":  50.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "strconv.ParseInt: parsing \"1k\": invalid syntax"},
		},
		{
			name: "invalid amount",
			id:   "1",
			requestBody: map[string]interface{}{
				"op_type": "withdraw",
				"amount":  "50.0",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "json: cannot unmarshal string into Go struct field .amount of type float64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request body
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create a new HTTP request with the test route and request body
			req, _ := http.NewRequest("PUT", "/api/balance/"+tt.id, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			var responseBody map[string]interface{}
			_ = json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}

type MockWalletUpdateErrRepo struct{}

func (m *MockWalletUpdateErrRepo) UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error {
	if walletID == 1 {
		return errors.New("update balance failed")
	}
	return errors.New("update balance failed")
}

func (m *MockWalletUpdateErrRepo) ExecTransfer(db *sql.DB, fromWalletID, toWalletID int64, amount float64) error {
	return nil
}

func (m *MockWalletUpdateErrRepo) GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error) {
	if walletID == 1 {
		return &Wallet{ID: 1, Balance: 100.0}, nil
	}
	return nil, errors.New("wallet not found")
}
func (m *MockWalletUpdateErrRepo) GetTransactionsByWalletID(db *sql.DB, walletID int64, limit, offset int) ([]Transaction, error) {
	if walletID == 1 {
		return []Transaction{
			{ID: 1, WalletID: 1, Amount: 50.0, OpType: "deposit"},
			{ID: 2, WalletID: 1, Amount: -20.0, OpType: "withdraw"},
		}, nil
	}
	return nil, errors.New("wallet not found")
}

func TestDepositWithdrawHandler_Err(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletUpdateErrRepo{}}
	router.PUT("/api/balance/:id", a.depositWithdrawHandler)

	// Test cases
	tests := []struct {
		name           string
		id             string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Update failed",
			id:   "1",
			requestBody: map[string]interface{}{
				"op_type": "deposit",
				"amount":  100.0,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "update balance failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request body
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create a new HTTP request with the test route and request body
			req, _ := http.NewRequest("PUT", "/api/balance/"+tt.id, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			var responseBody map[string]interface{}
			_ = json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}

type MockWalletTransferErrRepo struct{}

func (m *MockWalletTransferErrRepo) UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error {
	if walletID == 1 {
		return nil
	}
	return errors.New("update balance failed")
}

func (m *MockWalletTransferErrRepo) ExecTransfer(db *sql.DB, fromWalletID, toWalletID int64, amount float64) error {
	return fmt.Errorf("transfer failed: from %d to %d", fromWalletID, toWalletID)
}

func (m *MockWalletTransferErrRepo) GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error) {
	if walletID == 1 {
		return &Wallet{ID: 1, Balance: 100.0}, nil
	}
	return nil, errors.New("wallet not found")
}
func (m *MockWalletTransferErrRepo) GetTransactionsByWalletID(db *sql.DB, walletID int64, limit, offset int) ([]Transaction, error) {
	if walletID == 1 {
		return []Transaction{
			{ID: 1, WalletID: 1, Amount: 50.0, OpType: "deposit"},
			{ID: 2, WalletID: 1, Amount: -20.0, OpType: "withdraw"},
		}, nil
	}
	return nil, errors.New("wallet not found")
}

func TestTransferHandlerError(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletTransferErrRepo{}}
	router.POST("/api/transfer", a.transferHandler)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Transfer Error",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   2,
				"amount":         50.0,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "transfer failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request body
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create a new HTTP request with the test route and request body
			req, _ := http.NewRequest("POST", "/api/transfer", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			var responseBody map[string]interface{}
			_ = json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}

func TestTransferHandler(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletRepo{}}
	router.POST("/api/transfer", a.transferHandler)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Transfer Success",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   2,
				"amount":         50.0,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "transfer successful"},
		},
		{
			name: "Invalid Amount",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   2,
				"amount":         -50.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "transfer amount must be positive"},
		},
		{
			name: "Invalid Wallet ID",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   1,
				"amount":         50.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "invalid wallet id"},
		},
		{
			name: "From Wallet Not Found",
			requestBody: map[string]interface{}{
				"from_wallet_id": 999,
				"to_wallet_id":   2,
				"amount":         50.0,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "from wallet not found"},
		},
		{
			name: "Not Enough Balance",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   2,
				"amount":         150.0,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"error": "not enough"},
		},
		{
			name: "invalid amount",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   2,
				"amount":         "150k.0",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "json: cannot unmarshal string into Go struct field .amount of type float64"},
		},
		{
			name: "invalid amount",
			requestBody: map[string]interface{}{
				"from_wallet_id": 1,
				"to_wallet_id":   2,
				"amount":         "150k.0",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "json: cannot unmarshal string into Go struct field .amount of type float64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request body
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create a new HTTP request with the test route and request body
			req, _ := http.NewRequest("POST", "/api/transfer", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			var responseBody map[string]interface{}
			_ = json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}
func TestGetBalanceHandler(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletRepo{}}
	router.GET("/api/balance/:id", a.getBalanceHandler)

	// Test cases
	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Get Balance Success",
			id:             "1",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"balance": 100.0},
		},
		{
			name:           "Wallet Not Found",
			id:             "999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "wallet not found"},
		},
		{
			name:           "Invalid Wallet ID",
			id:             "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "strconv.ParseInt: parsing \"invalid\": invalid syntax"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with the test route
			req, _ := http.NewRequest("GET", "/api/balance/"+tt.id, nil)

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			var responseBody map[string]interface{}
			_ = json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}

type MockWalletGetTransactionErrRepo struct{}

func (m *MockWalletGetTransactionErrRepo) UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error {
	if walletID == 1 {
		return errors.New("update balance failed")
	}
	return nil
}

func (m *MockWalletGetTransactionErrRepo) ExecTransfer(db *sql.DB, fromWalletID, toWalletID int64, amount float64) error {
	return nil
}

func (m *MockWalletGetTransactionErrRepo) GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error) {
	if walletID == 1 {
		return &Wallet{ID: 1, Balance: 100.0}, nil
	}
	return nil, errors.New("wallet not found")
}
func (m *MockWalletGetTransactionErrRepo) GetTransactionsByWalletID(db *sql.DB, walletID int64, limit, offset int) ([]Transaction, error) {
	if walletID == 1 {
		return []Transaction{
			{ID: 1, WalletID: 1, Amount: 50.0, OpType: "deposit"},
			{ID: 2, WalletID: 1, Amount: -20.0, OpType: "withdraw"},
		}, errors.New("db error")
	}
	return nil, errors.New("db error")
}

func TestGetTransactionsHandler_Error(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletGetTransactionErrRepo{}}
	router.GET("/api/transaction/:id", a.getTransactions)

	// Test cases
	tests := []struct {
		name           string
		id             string
		limit          string
		offset         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Get Transactions Error",
			id:             "1",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "db error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with the test route
			req, _ := http.NewRequest("GET", "/api/transaction/"+tt.id+"?limit="+tt.limit+"&offset="+tt.offset, nil)

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			//var responseBody map[string]interface{}

			expected, _ := json.Marshal(tt.expectedBody)
			assert.Equal(t, string(expected), rec.Body.String())
		})
	}
}

type MockWalletGetTransactionWalletErrRepo struct{}

func (m *MockWalletGetTransactionWalletErrRepo) UpdateBalance(db *sql.DB, walletID int64, opType string, amount float64) error {
	if walletID == 1 {
		return errors.New("update balance failed")
	}
	return nil
}

func (m *MockWalletGetTransactionWalletErrRepo) ExecTransfer(db *sql.DB, fromWalletID, toWalletID int64, amount float64) error {
	return nil
}

func (m *MockWalletGetTransactionWalletErrRepo) GetWalletInfoById(db *sql.DB, walletID int64) (*Wallet, error) {
	if walletID == 1 {
		return &Wallet{ID: 1, Balance: 100.0}, errors.New("db error")
	}
	return nil, errors.New("db error")
}
func (m *MockWalletGetTransactionWalletErrRepo) GetTransactionsByWalletID(db *sql.DB, walletID int64, limit, offset int) ([]Transaction, error) {
	if walletID == 1 {
		return []Transaction{
			{ID: 1, WalletID: 1, Amount: 50.0, OpType: "deposit"},
			{ID: 2, WalletID: 1, Amount: -20.0, OpType: "withdraw"},
		}, errors.New("db error")
	}
	return nil, errors.New("db error")
}

func TestGetTransactionsHandler_WalletError(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletGetTransactionWalletErrRepo{}}
	router.GET("/api/transaction/:id", a.getTransactions)

	// Test cases
	tests := []struct {
		name           string
		id             string
		limit          string
		offset         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Get Transactions Error",
			id:             "1",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "db error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with the test route
			req, _ := http.NewRequest("GET", "/api/transaction/"+tt.id+"?limit="+tt.limit+"&offset="+tt.offset, nil)

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			//var responseBody map[string]interface{}

			expected, _ := json.Marshal(tt.expectedBody)
			assert.Equal(t, string(expected), rec.Body.String())
		})
	}
}

func (m *MockWalletRepo) GetTransactionsByWalletID(db *sql.DB, walletID int64, limit, offset int) ([]Transaction, error) {
	if walletID == 1 {
		return []Transaction{
			{ID: 1, WalletID: 1, Amount: 50.0, OpType: "deposit"},
			{ID: 2, WalletID: 1, Amount: -20.0, OpType: "withdraw"},
		}, nil
	}
	return nil, errors.New("wallet not found")
}

func TestGetTransactionsHandler(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Initialize the app and set up the route
	a := App{Rp: &MockWalletRepo{}}
	router.GET("/api/transaction/:id", a.getTransactions)

	// Test cases
	tests := []struct {
		name           string
		id             string
		limit          string
		offset         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Get Transactions Success",
			id:             "1",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"transactions": []Transaction{
					{ID: 1, WalletID: 1, Amount: 50.0, OpType: "deposit"},
					{ID: 2, WalletID: 1, Amount: -20.0, OpType: "withdraw"},
				},
			},
		},
		{
			name:           "Wallet Not Found",
			id:             "999",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "wallet not found"},
		},
		{
			name:           "Invalid Limit",
			id:             "1",
			limit:          "-1",
			offset:         "0",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "invalid limit"},
		},
		{
			name:           "Invalid Offset",
			id:             "1",
			limit:          "10",
			offset:         "-1",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "invalid offset"},
		},
		{
			name:           "Invalid Wallet ID",
			id:             "invalid",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "strconv.ParseInt: parsing \"invalid\": invalid syntax"},
		},
		{
			name:           "Invalid Wallet ID",
			id:             "invalid",
			limit:          "10",
			offset:         "0",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "strconv.ParseInt: parsing \"invalid\": invalid syntax"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with the test route
			req, _ := http.NewRequest("GET", "/api/transaction/"+tt.id+"?limit="+tt.limit+"&offset="+tt.offset, nil)

			// Create a new HTTP response recorder
			rec := httptest.NewRecorder()

			// Perform the HTTP request
			router.ServeHTTP(rec, req)

			// Assert that the response status code is as expected
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert that the response body is as expected
			//var responseBody map[string]interface{}

			expected, _ := json.Marshal(tt.expectedBody)
			assert.Equal(t, string(expected), rec.Body.String())
		})
	}
}
