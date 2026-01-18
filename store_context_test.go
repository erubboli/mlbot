package main

import (
	"context"
	"errors"
	"testing"
)

func TestSQLStoreRespectsContextCancel(t *testing.T) {
	db, err := initDB(":memory:")
	if err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	store, err := NewSQLStore(db)
	if err != nil {
		t.Fatalf("NewSQLStore failed: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = store.GetPools(ctx, "user-1")
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
