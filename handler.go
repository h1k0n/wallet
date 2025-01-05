package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountRequest struct {
	Id int64 `uri:"id" binding:"required"`
}

func (a *App) depositWithdrawHandler(c *gin.Context) {
	var req AccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var request struct {
		OpType string  `json:"op_type"`
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !(request.OpType == "deposit" || request.OpType == "withdraw") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation type"})
		return
	}
	// 获取钱包信息
	wallet, err := a.Rp.GetWalletInfoById(a.DB, req.Id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	if request.OpType == "withdraw" && wallet.Balance < request.Amount {
		c.JSON(http.StatusOK, gin.H{"message": "not enough"})
		return

	}
	if request.OpType == "withdraw" {
		request.Amount = -request.Amount
	}
	err = a.Rp.UpdateBalance(a.DB, wallet.ID, request.OpType, request.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": request.OpType + " successful"})
}

func (a *App) transferHandler(c *gin.Context) {
	var request struct {
		FromWalletId int64   `json:"from_wallet_id"`
		ToWalletId   int64   `json:"to_wallet_id"`
		Amount       float64 `json:"amount"`
	}

	// 解析请求参数
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查金额是否为正数
	if request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer amount must be positive"})
		return
	}

	if request.FromWalletId <= 0 || request.ToWalletId <= 0 || request.FromWalletId == request.ToWalletId {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}

	// 获取发起钱包和接收钱包的信息
	fromWallet, err := a.Rp.GetWalletInfoById(a.DB, request.FromWalletId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "from wallet not found"})
		return
	}
	if fromWallet.Balance < request.Amount {
		c.JSON(http.StatusOK, gin.H{"error": "not enough"})
		return
	}

	a.Rp.ExecTransfer(a.DB, request.FromWalletId, request.ToWalletId, request.Amount)
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func (a *App) getBalanceHandler(c *gin.Context) {
	var req AccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取钱包信息
	wallet, err := a.Rp.GetWalletInfoById(a.DB, req.Id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": wallet.Balance})
}

// 查询交易记录的 Handler
func (a *App) getTransactions(c *gin.Context) {
	var req AccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit := c.DefaultQuery("limit", "10")  // 每页默认 10 条记录
	offset := c.DefaultQuery("offset", "0") // 默认从第 0 条记录开始

	// 将查询参数转换为整数
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
		return
	}

	// 获取钱包信息
	wallet, err := a.Rp.GetWalletInfoById(a.DB, req.Id)
	if err != nil {
		if err.Error() == "wallet not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 获取钱包的交易记录
	transactions, err := a.Rp.GetTransactionsByWalletID(a.DB, wallet.ID, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回交易记录
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}
