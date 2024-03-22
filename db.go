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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		chatID TEXT NOT NULL,
		UNIQUE(userID, chatID) ON CONFLICT IGNORE
	)`)
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS delegations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		delegationID TEXT NOT NULL,
		balance INTEGER DEFAULT 0,
		UNIQUE(userID, delegationID) ON CONFLICT IGNORE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS addresses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		address TEXT NOT NULL,
		balance INTEGER DEFAULT 0,
		UNIQUE(userID, address) ON CONFLICT IGNORE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

type Notification struct {
	UserID string
	ChatID int64
}

func getAllNotifications(db *sql.DB) ([]Notification, error) {
	rows, err := db.Query("SELECT userID, chatID FROM notifications")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(&notification.UserID, &notification.ChatID); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func addNotification(db *sql.DB, userID string, chatID int64) error {
	_, err := db.Exec("INSERT INTO notifications (userID, chatID) VALUES (?, ?)", userID, chatID)
	return err
}

func removeNotification(db *sql.DB, userID string, chatID int64) error {
	_, err := db.Exec("DELETE FROM notifications WHERE userID = ? AND chatID = ?", userID, chatID)
	return err
}

func addDelegation(db *sql.DB, userID, delegationID string) error {
	hrp, _, err := bech32.Decode(delegationID)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	if hrp != "mdelg" {
		fmt.Println("Error: this is not a delegation mainnet address")
		return errors.New("this is not a delegation mainnet address")
	}

	_, err = db.Exec("INSERT INTO delegations (userID, delegationID, balance) VALUES (?, ?, 0)", userID, delegationID)
	return err
}

func updateDelegationBalance(db *sql.DB, userID, delegationID string, balance int64) error {
	_, err := db.Exec("UPDATE delegations SET balance = ? WHERE delegationID = ? AND userID = ?", balance, delegationID, userID)
	return err
}

func removeDelegation(db *sql.DB, userID, delegationID string) error {
	_, err := db.Exec("DELETE FROM delegations WHERE userID = ? AND delegationID = ?", userID, delegationID)
	return err
}

func getDelegationBalanceFromDb(db *sql.DB, userID, delegationID string) (int64, error) {
	var balance int64
	err := db.QueryRow("SELECT balance FROM delegations WHERE userID = ? AND delegationID = ?", userID, delegationID).Scan(&balance)
	return balance, err
}

func getDelegations(db *sql.DB, userID string) ([]string, error) {
	rows, err := db.Query("SELECT delegationID FROM delegations WHERE userID = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var delegations []string
	for rows.Next() {
		var delegationID string
		if err := rows.Scan(&delegationID); err != nil {
			return nil, err
		}
		delegations = append(delegations, delegationID)
	}
	return delegations, nil
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
