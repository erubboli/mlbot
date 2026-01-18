package main

import "testing"

func TestNewSQLStoreReturnsErrorOnClosedDB(t *testing.T) {
	db, err := initDB(":memory:")
	if err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	_ = db.Close()

	store, err := NewSQLStore(db)
	if err == nil {
		if store != nil {
			_ = store.Close()
		}
		t.Fatal("expected error when preparing statements on closed DB")
	}
}
