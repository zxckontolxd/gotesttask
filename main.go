package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB

type WalletOperation struct {
	WalletID     uuid.UUID `json:"walletId"`
	OperationType string   `json:"operationType"`
	Amount       int64     `json:"amount"`
}

func main() {
	var err error
	// изменить параметры подключения
	// не могу сейчас пофиксить под докер, так как wsl умер окончательно и нужно переустановить систему (бесконечно долго устанавливает убунту, не пишет лог)
	// до этого код запускался, но не подключался к бд, пока я не сделал глупость, от чего весь wsl поломался
	connStr := "postgres://postgres:@localhost:5432/wallet_db?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/api/v1/wallet", handleWalletOperation)
	http.HandleFunc("/api/v1/wallets/", getWalletBalance)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var op WalletOperation
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if op.OperationType != "DEPOSIT" && op.OperationType != "WITHDRAW" {
		http.Error(w, "Invalid operation type", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	var balance int64
	err = tx.QueryRow("SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", op.WalletID).Scan(&balance)
	if err == sql.ErrNoRows {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		tx.Rollback()
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve balance", http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	if op.OperationType == "WITHDRAW" && balance < op.Amount {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		tx.Rollback()
		return
	}

	var newBalance int64
	if op.OperationType == "DEPOSIT" {
		newBalance = balance + op.Amount
	} else if op.OperationType == "WITHDRAW" {
		newBalance = balance - op.Amount
	}

	_, err = tx.Exec("UPDATE wallets SET balance = $1 WHERE id = $2", newBalance, op.WalletID)
	if err != nil {
		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Wallet updated successfully")
}

func getWalletBalance(w http.ResponseWriter, r *http.Request) {
	walletIDStr := r.URL.Path[len("/api/v1/wallets/"):]
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	var balance int64
	err = db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	if err == sql.ErrNoRows {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	 } else if err != nil {
		http.Error(w, "Failed to retrieve balance", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Wallet balance: %d", balance)
}