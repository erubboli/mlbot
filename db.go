package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil/bech32"
	_ "github.com/mattn/go-sqlite3"
)

func initDB(filePath string) *sql.DB {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pools (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		poolID TEXT NOT NULL,
		balance INTEGER DEFAULT 0,
		UNIQUE(userID, poolID) ON CONFLICT IGNORE
	)`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func addPool(db *sql.DB, userID, poolID string) error {
	hrp, _, err := bech32.Decode(poolID)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	if hrp != "mpool" {
		fmt.Println("Error: this is not a pool mainnet address")
		return errors.New("this is not a pool mainnet address")
	}

	_, err = db.Exec("INSERT INTO pools (userID, poolID, balance) VALUES (?, ?, 0)", userID, poolID)
	return err
}

func updatePoolBalance(db *sql.DB, userID, poolID string, balance int64) error {
	_, err := db.Exec("UPDATE pools SET balance = ? WHERE poolID = ? AND userID = ?", balance, poolID, userID)
	return err
}

func removePool(db *sql.DB, userID, poolID string) error {
	_, err := db.Exec("DELETE FROM pools WHERE userID = ? AND poolID = ?", userID, poolID)
	return err
}

func getPoolBalanceFromDb(db *sql.DB, userID, poolID string) (int64, error) {
	var balance int64
	err := db.QueryRow("SELECT balance FROM pools WHERE userID = ? AND poolID = ?", userID, poolID).Scan(&balance)
	return balance, err
}

func getPools(db *sql.DB, userID string) ([]string, error) {
	rows, err := db.Query("SELECT poolID FROM pools WHERE userID = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []string
	for rows.Next() {
		var poolID string
		if err := rows.Scan(&poolID); err != nil {
			return nil, err
		}
		pools = append(pools, poolID)
	}
	return pools, nil
}
