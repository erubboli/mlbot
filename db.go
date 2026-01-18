package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/bech32"
	_ "github.com/mattn/go-sqlite3"
)

func initDB(filePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	if err := applyDBPragmas(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		chatID TEXT NOT NULL,
		UNIQUE(userID, chatID) ON CONFLICT IGNORE
	)`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pools (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		poolID TEXT NOT NULL,
		balance INTEGER DEFAULT 0,
		UNIQUE(userID, poolID) ON CONFLICT IGNORE
	)`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS delegations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		delegationID TEXT NOT NULL,
		balance INTEGER DEFAULT 0,
		UNIQUE(userID, delegationID) ON CONFLICT IGNORE
	)`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS addresses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userID TEXT NOT NULL,
		address TEXT NOT NULL,
		balance INTEGER DEFAULT 0,
		notify_on_change BOOLEAN DEFAULT FALSE,
		threshold INT,		
		UNIQUE(userID, address) ON CONFLICT IGNORE
	)`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func applyDBPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}
	return nil
}

type Notification struct {
	UserID string
	ChatID int64
}

func getAllNotificationsWithContext(ctx context.Context, db *sql.DB) ([]Notification, error) {
	rows, err := db.QueryContext(ctx, "SELECT userID, chatID FROM notifications")
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

func addMonitoredAddressWithContext(ctx context.Context, db *sql.DB, userID, address string, threshold int, notifyOnChange bool, chatID int64) error {
	_, err := db.ExecContext(ctx, "INSERT INTO addresses (userID, address, threshold, notify_on_change) VALUES (?, ?, ?, ?)", userID, address, threshold, notifyOnChange)
	if err != nil {
		return err
	}
	if notifyOnChange {
		err = addNotificationWithContext(ctx, db, userID, chatID)
	}
	return err
}

func removeMonitoredAddressWithContext(ctx context.Context, db *sql.DB, userID, address string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM addresses WHERE userID = ? AND address = ?", userID, address)
	return err
}

func updateAddressBalanceWithContext(ctx context.Context, db *sql.DB, userID, address string, balance int64) error {
	_, err := db.ExecContext(ctx, "UPDATE addresses SET balance = ? WHERE address = ? AND userID = ?", balance, address, userID)
	return err
}

func addNotificationWithContext(ctx context.Context, db *sql.DB, userID string, chatID int64) error {
	_, err := db.ExecContext(ctx, "INSERT INTO notifications (userID, chatID) VALUES (?, ?)", userID, chatID)
	return err
}

func removeNotificationWithContext(ctx context.Context, db *sql.DB, userID string, chatID int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM notifications WHERE userID = ? AND chatID = ?", userID, chatID)
	return err
}

func addDelegationWithContext(ctx context.Context, db *sql.DB, userID, delegationID string) error {
	hrp, _, err := bech32.Decode(delegationID)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	if hrp != "mdelg" {
		fmt.Println("Error: this is not a delegation mainnet address")
		return errors.New("this is not a delegation mainnet address")
	}

	_, err = db.ExecContext(ctx, "INSERT INTO delegations (userID, delegationID, balance) VALUES (?, ?, 0)", userID, delegationID)
	return err
}

func updateDelegationBalanceWithContext(ctx context.Context, db *sql.DB, userID, delegationID string, balance int64) error {
	_, err := db.ExecContext(ctx, "UPDATE delegations SET balance = ? WHERE delegationID = ? AND userID = ?", balance, delegationID, userID)
	return err
}

func removeDelegationWithContext(ctx context.Context, db *sql.DB, userID, delegationID string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM delegations WHERE userID = ? AND delegationID = ?", userID, delegationID)
	return err
}

func getDelegationBalanceFromDbWithContext(ctx context.Context, db *sql.DB, userID, delegationID string) (int64, error) {
	var balance int64
	err := db.QueryRowContext(ctx, "SELECT balance FROM delegations WHERE userID = ? AND delegationID = ?", userID, delegationID).Scan(&balance)
	return balance, err
}

func getDelegationsWithContext(ctx context.Context, db *sql.DB, userID string) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT delegationID FROM delegations WHERE userID = ?", userID)
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

func addPoolWithContext(ctx context.Context, db *sql.DB, userID, poolID string) error {
	hrp, _, err := bech32.Decode(poolID)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	if hrp != "mpool" {
		fmt.Println("Error: this is not a pool mainnet address")
		return errors.New("this is not a pool mainnet address")
	}

	_, err = db.ExecContext(ctx, "INSERT INTO pools (userID, poolID, balance) VALUES (?, ?, 0)", userID, poolID)
	return err
}

func updatePoolBalanceWithContext(ctx context.Context, db *sql.DB, userID, poolID string, balance int64) error {
	_, err := db.ExecContext(ctx, "UPDATE pools SET balance = ? WHERE poolID = ? AND userID = ?", balance, poolID, userID)
	return err
}

func removePoolWithContext(ctx context.Context, db *sql.DB, userID, poolID string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM pools WHERE userID = ? AND poolID = ?", userID, poolID)
	return err
}

func getPoolBalanceFromDbWithContext(ctx context.Context, db *sql.DB, userID, poolID string) (int64, error) {
	var balance int64
	err := db.QueryRowContext(ctx, "SELECT balance FROM pools WHERE userID = ? AND poolID = ?", userID, poolID).Scan(&balance)
	return balance, err
}

func getPoolsWithContext(ctx context.Context, db *sql.DB, userID string) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT poolID FROM pools WHERE userID = ?", userID)
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
