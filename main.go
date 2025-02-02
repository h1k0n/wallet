package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

type App struct {
	DB *sql.DB
	Rp IWallet
}

func main() {
	a := App{}
	a.initDB(os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_HOST"))

	r := gin.Default()
	r.PUT("/api/balance/:id", a.depositWithdrawHandler) //deposit and withdraw
	r.GET("/api/balance/:id", a.getBalanceHandler)
	r.POST("/api/transfer", a.transferHandler)
	r.GET("/api/transaction/:id", a.getTransactions)

	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
