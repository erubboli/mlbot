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
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) AddMonitoredAddress(ctx context.Context, userID, address string, threshold int, notifyOnChange bool, chatID int64) error {
	return addMonitoredAddressWithContext(ctx, s.db, userID, address, threshold, notifyOnChange, chatID)
}

func (s *SQLStore) AddPool(ctx context.Context, userID, poolID string) error {
	return addPoolWithContext(ctx, s.db, userID, poolID)
}

func (s *SQLStore) RemovePool(ctx context.Context, userID, poolID string) error {
	return removePoolWithContext(ctx, s.db, userID, poolID)
}

func (s *SQLStore) GetPools(ctx context.Context, userID string) ([]string, error) {
	return getPoolsWithContext(ctx, s.db, userID)
}

func (s *SQLStore) AddDelegation(ctx context.Context, userID, delegationID string) error {
	return addDelegationWithContext(ctx, s.db, userID, delegationID)
}

func (s *SQLStore) RemoveDelegation(ctx context.Context, userID, delegationID string) error {
	return removeDelegationWithContext(ctx, s.db, userID, delegationID)
}

func (s *SQLStore) GetDelegations(ctx context.Context, userID string) ([]string, error) {
	return getDelegationsWithContext(ctx, s.db, userID)
}

func (s *SQLStore) GetPoolBalance(ctx context.Context, userID, poolID string) (int64, error) {
	return getPoolBalanceFromDbWithContext(ctx, s.db, userID, poolID)
}

func (s *SQLStore) UpdatePoolBalance(ctx context.Context, userID, poolID string, balance int64) error {
	return updatePoolBalanceWithContext(ctx, s.db, userID, poolID, balance)
}

func (s *SQLStore) GetDelegationBalance(ctx context.Context, userID, delegationID string) (int64, error) {
	return getDelegationBalanceFromDbWithContext(ctx, s.db, userID, delegationID)
}

func (s *SQLStore) UpdateDelegationBalance(ctx context.Context, userID, delegationID string, balance int64) error {
	return updateDelegationBalanceWithContext(ctx, s.db, userID, delegationID, balance)
}

func (s *SQLStore) AddNotification(ctx context.Context, userID string, chatID int64) error {
	return addNotificationWithContext(ctx, s.db, userID, chatID)
}

func (s *SQLStore) RemoveNotification(ctx context.Context, userID string, chatID int64) error {
	return removeNotificationWithContext(ctx, s.db, userID, chatID)
}

func (s *SQLStore) GetAllNotifications(ctx context.Context) ([]Notification, error) {
	return getAllNotificationsWithContext(ctx, s.db)
}
