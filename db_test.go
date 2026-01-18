package main

import "testing"

func TestInitDBInvalidPathReturnsError(t *testing.T) {
	db, err := initDB("/path/does/not/exist/pools.db")
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatal("expected error for invalid database path")
	}
}
