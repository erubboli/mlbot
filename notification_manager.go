package main

import (
	"context"
	"sync"
)

type NotificationManager struct {
	mu      sync.Mutex
	entries map[string]context.CancelFunc
	wg      sync.WaitGroup
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		entries: make(map[string]context.CancelFunc),
	}
}

func (m *NotificationManager) Start(ctx context.Context, userID string, start func(ctx context.Context)) bool {
	m.mu.Lock()
	if _, exists := m.entries[userID]; exists {
		m.mu.Unlock()
		return false
	}
	ctxIn, cancel := context.WithCancel(ctx)
	m.entries[userID] = cancel
	m.wg.Add(1)
	m.mu.Unlock()

	go func() {
		defer m.wg.Done()
		start(ctxIn)
	}()
	return true
}

func (m *NotificationManager) Active(userID string) bool {
	m.mu.Lock()
	_, exists := m.entries[userID]
	m.mu.Unlock()
	return exists
}

func (m *NotificationManager) Stop(userID string) bool {
	m.mu.Lock()
	cancel, exists := m.entries[userID]
	if exists {
		delete(m.entries, userID)
	}
	m.mu.Unlock()
	if exists {
		cancel()
	}
	return exists
}

func (m *NotificationManager) StopAll() {
	m.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(m.entries))
	for _, cancel := range m.entries {
		cancels = append(cancels, cancel)
	}
	m.entries = make(map[string]context.CancelFunc)
	m.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
	m.wg.Wait()
}
