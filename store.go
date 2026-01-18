package main

import (
	"context"
	"database/sql"
)

type Store interface {
	AddMonitoredAddress(ctx context.Context, userID, address string, threshold int, notifyOnChange bool, chatID int64) error
	AddPool(ctx context.Context, userID, poolID string) error
	RemovePool(ctx context.Context, userID, poolID string) error
	GetPools(ctx context.Context, userID string) ([]string, error)
	AddDelegation(ctx context.Context, userID, delegationID string) error
	RemoveDelegation(ctx context.Context, userID, delegationID string) error
	GetDelegations(ctx context.Context, userID string) ([]string, error)
	GetPoolBalance(ctx context.Context, userID, poolID string) (int64, error)
	UpdatePoolBalance(ctx context.Context, userID, poolID string, balance int64) error
	GetDelegationBalance(ctx context.Context, userID, delegationID string) (int64, error)
	UpdateDelegationBalance(ctx context.Context, userID, delegationID string, balance int64) error
	AddNotification(ctx context.Context, userID string, chatID int64) error
	RemoveNotification(ctx context.Context, userID string, chatID int64) error
	GetAllNotifications(ctx context.Context) ([]Notification, error)
}

type SQLStore struct {
	db *sql.DB

	stmtAddNotification         *sql.Stmt
	stmtRemoveNotification      *sql.Stmt
	stmtGetPools                *sql.Stmt
	stmtGetDelegations          *sql.Stmt
	stmtGetPoolBalance          *sql.Stmt
	stmtUpdatePoolBalance       *sql.Stmt
	stmtGetDelegationBalance    *sql.Stmt
	stmtUpdateDelegationBalance *sql.Stmt
}

func NewSQLStore(db *sql.DB) (*SQLStore, error) {
	store := &SQLStore{db: db}
	if err := store.prepareStatements(); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLStore) prepareStatements() error {
	var err error

	s.stmtAddNotification, err = s.db.Prepare("INSERT INTO notifications (userID, chatID) VALUES (?, ?)")
	if err != nil {
		return err
	}
	s.stmtRemoveNotification, err = s.db.Prepare("DELETE FROM notifications WHERE userID = ? AND chatID = ?")
	if err != nil {
		return err
	}
	s.stmtGetPools, err = s.db.Prepare("SELECT poolID FROM pools WHERE userID = ?")
	if err != nil {
		return err
	}
	s.stmtGetDelegations, err = s.db.Prepare("SELECT delegationID FROM delegations WHERE userID = ?")
	if err != nil {
		return err
	}
	s.stmtGetPoolBalance, err = s.db.Prepare("SELECT balance FROM pools WHERE userID = ? AND poolID = ?")
	if err != nil {
		return err
	}
	s.stmtUpdatePoolBalance, err = s.db.Prepare("UPDATE pools SET balance = ? WHERE poolID = ? AND userID = ?")
	if err != nil {
		return err
	}
	s.stmtGetDelegationBalance, err = s.db.Prepare("SELECT balance FROM delegations WHERE userID = ? AND delegationID = ?")
	if err != nil {
		return err
	}
	s.stmtUpdateDelegationBalance, err = s.db.Prepare("UPDATE delegations SET balance = ? WHERE delegationID = ? AND userID = ?")
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStore) Close() error {
	var firstErr error
	closeStmt := func(stmt *sql.Stmt) {
		if stmt == nil {
			return
		}
		if err := stmt.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	closeStmt(s.stmtAddNotification)
	closeStmt(s.stmtRemoveNotification)
	closeStmt(s.stmtGetPools)
	closeStmt(s.stmtGetDelegations)
	closeStmt(s.stmtGetPoolBalance)
	closeStmt(s.stmtUpdatePoolBalance)
	closeStmt(s.stmtGetDelegationBalance)
	closeStmt(s.stmtUpdateDelegationBalance)
	return firstErr
}

func (s *SQLStore) AddMonitoredAddress(ctx context.Context, userID, address string, threshold int, notifyOnChange bool, chatID int64) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO addresses (userID, address, threshold, notify_on_change) VALUES (?, ?, ?, ?)", userID, address, threshold, notifyOnChange)
	if err != nil {
		return err
	}
	if notifyOnChange {
		return s.AddNotification(ctx, userID, chatID)
	}
	return nil
}

func (s *SQLStore) AddPool(ctx context.Context, userID, poolID string) error {
	return addPoolWithContext(ctx, s.db, userID, poolID)
}

func (s *SQLStore) RemovePool(ctx context.Context, userID, poolID string) error {
	return removePoolWithContext(ctx, s.db, userID, poolID)
}

func (s *SQLStore) GetPools(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.stmtGetPools.QueryContext(ctx, userID)
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

func (s *SQLStore) AddDelegation(ctx context.Context, userID, delegationID string) error {
	return addDelegationWithContext(ctx, s.db, userID, delegationID)
}

func (s *SQLStore) RemoveDelegation(ctx context.Context, userID, delegationID string) error {
	return removeDelegationWithContext(ctx, s.db, userID, delegationID)
}

func (s *SQLStore) GetDelegations(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.stmtGetDelegations.QueryContext(ctx, userID)
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

func (s *SQLStore) GetPoolBalance(ctx context.Context, userID, poolID string) (int64, error) {
	var balance int64
	err := s.stmtGetPoolBalance.QueryRowContext(ctx, userID, poolID).Scan(&balance)
	return balance, err
}

func (s *SQLStore) UpdatePoolBalance(ctx context.Context, userID, poolID string, balance int64) error {
	_, err := s.stmtUpdatePoolBalance.ExecContext(ctx, balance, poolID, userID)
	return err
}

func (s *SQLStore) GetDelegationBalance(ctx context.Context, userID, delegationID string) (int64, error) {
	var balance int64
	err := s.stmtGetDelegationBalance.QueryRowContext(ctx, userID, delegationID).Scan(&balance)
	return balance, err
}

func (s *SQLStore) UpdateDelegationBalance(ctx context.Context, userID, delegationID string, balance int64) error {
	_, err := s.stmtUpdateDelegationBalance.ExecContext(ctx, balance, delegationID, userID)
	return err
}

func (s *SQLStore) AddNotification(ctx context.Context, userID string, chatID int64) error {
	_, err := s.stmtAddNotification.ExecContext(ctx, userID, chatID)
	return err
}

func (s *SQLStore) RemoveNotification(ctx context.Context, userID string, chatID int64) error {
	_, err := s.stmtRemoveNotification.ExecContext(ctx, userID, chatID)
	return err
}

func (s *SQLStore) GetAllNotifications(ctx context.Context) ([]Notification, error) {
	return getAllNotificationsWithContext(ctx, s.db)
}
