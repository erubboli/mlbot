package main

import "database/sql"

type Store interface {
	AddMonitoredAddress(userID, address string, threshold int, notifyOnChange bool, chatID int64) error
	AddPool(userID, poolID string) error
	RemovePool(userID, poolID string) error
	GetPools(userID string) ([]string, error)
	AddDelegation(userID, delegationID string) error
	RemoveDelegation(userID, delegationID string) error
	GetDelegations(userID string) ([]string, error)
	GetPoolBalance(userID, poolID string) (int64, error)
	UpdatePoolBalance(userID, poolID string, balance int64) error
	GetDelegationBalance(userID, delegationID string) (int64, error)
	UpdateDelegationBalance(userID, delegationID string, balance int64) error
	AddNotification(userID string, chatID int64) error
	RemoveNotification(userID string, chatID int64) error
	GetAllNotifications() ([]Notification, error)
}

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) AddMonitoredAddress(userID, address string, threshold int, notifyOnChange bool, chatID int64) error {
	return addMonitoredAddress(s.db, userID, address, threshold, notifyOnChange, chatID)
}

func (s *SQLStore) AddPool(userID, poolID string) error {
	return addPool(s.db, userID, poolID)
}

func (s *SQLStore) RemovePool(userID, poolID string) error {
	return removePool(s.db, userID, poolID)
}

func (s *SQLStore) GetPools(userID string) ([]string, error) {
	return getPools(s.db, userID)
}

func (s *SQLStore) AddDelegation(userID, delegationID string) error {
	return addDelegation(s.db, userID, delegationID)
}

func (s *SQLStore) RemoveDelegation(userID, delegationID string) error {
	return removeDelegation(s.db, userID, delegationID)
}

func (s *SQLStore) GetDelegations(userID string) ([]string, error) {
	return getDelegations(s.db, userID)
}

func (s *SQLStore) GetPoolBalance(userID, poolID string) (int64, error) {
	return getPoolBalanceFromDb(s.db, userID, poolID)
}

func (s *SQLStore) UpdatePoolBalance(userID, poolID string, balance int64) error {
	return updatePoolBalance(s.db, userID, poolID, balance)
}

func (s *SQLStore) GetDelegationBalance(userID, delegationID string) (int64, error) {
	return getDelegationBalanceFromDb(s.db, userID, delegationID)
}

func (s *SQLStore) UpdateDelegationBalance(userID, delegationID string, balance int64) error {
	return updateDelegationBalance(s.db, userID, delegationID, balance)
}

func (s *SQLStore) AddNotification(userID string, chatID int64) error {
	return addNotification(s.db, userID, chatID)
}

func (s *SQLStore) RemoveNotification(userID string, chatID int64) error {
	return removeNotification(s.db, userID, chatID)
}

func (s *SQLStore) GetAllNotifications() ([]Notification, error) {
	return getAllNotifications(s.db)
}
