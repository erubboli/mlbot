package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNotificationManagerStartStopAll(t *testing.T) {
	manager := NewNotificationManager()

	started := make(chan struct{})
	stopped := make(chan struct{})

	ok := manager.Start(context.Background(), "user-1", func(ctx context.Context) {
		close(started)
		<-ctx.Done()
		close(stopped)
	})
	if !ok {
		t.Fatal("expected first start to succeed")
	}

	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for routine to start")
	}

	done := make(chan struct{})
	go func() {
		manager.StopAll()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for stop all")
	}

	select {
	case <-stopped:
	default:
		t.Fatal("expected routine to stop on StopAll")
	}
}

func TestNotificationManagerStartDuplicate(t *testing.T) {
	manager := NewNotificationManager()
	var starts int32

	ok := manager.Start(context.Background(), "user-1", func(ctx context.Context) {
		atomic.AddInt32(&starts, 1)
		<-ctx.Done()
	})
	if !ok {
		t.Fatal("expected first start to succeed")
	}

	ok = manager.Start(context.Background(), "user-1", func(ctx context.Context) {
		atomic.AddInt32(&starts, 1)
		<-ctx.Done()
	})
	if ok {
		t.Fatal("expected duplicate start to be rejected")
	}

	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&starts) != 1 {
		t.Fatalf("expected 1 routine start, got %d", atomic.LoadInt32(&starts))
	}

	manager.StopAll()
}
