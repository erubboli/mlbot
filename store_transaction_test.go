package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestAddMonitoredAddressTransactional(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tx.db")
	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	store, err := NewSQLStore(db)
	if err != nil {
		t.Fatalf("NewSQLStore failed: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if _, err := db.Exec("DROP TABLE notifications"); err != nil {
		t.Fatalf("failed to drop notifications: %v", err)
	}

	ctx := context.Background()
	err = store.AddMonitoredAddress(ctx, "user-1", "addr-1", 0, true, 10)
	if err == nil {
		t.Fatal("expected error when notifications table is missing")
	}

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM addresses WHERE userID = ? AND address = ?", "user-1", "addr-1").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query addresses: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no addresses inserted after rollback, got %d", count)
	}
}
