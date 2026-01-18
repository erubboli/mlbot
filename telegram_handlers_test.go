package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TestNotifyStatusHandler(t *testing.T) {
	store := &fakeStore{}
	client := &noopBalanceClient{}
	app := NewApp(store, client, nil, NewNotificationManager(), "42")

	var lastMessage string
	app.send = func(ctx context.Context, _ *bot.Bot, _ int64, message string) {
		lastMessage = message
	}

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 1},
			From: &models.User{ID: 42},
		},
	}

	app.notifyStatusHandler(context.Background(), nil, update)
	if lastMessage != "Not Subscribed" {
		t.Fatalf("expected 'Not Subscribed', got %q", lastMessage)
	}

	app.notify.Start(context.Background(), "42", func(ctx context.Context) {
		<-ctx.Done()
	})
	t.Cleanup(app.notify.StopAll)

	app.notifyStatusHandler(context.Background(), nil, update)
	if lastMessage != "Subscribed" {
		t.Fatalf("expected 'Subscribed', got %q", lastMessage)
	}
}

func TestNotifyStartHandler(t *testing.T) {
	store := &fakeStore{}
	client := &noopBalanceClient{}
	app := NewApp(store, client, nil, NewNotificationManager(), "99")

	started := make(chan struct{}, 2)
	var starts int32
	app.startNotify = func(ctx context.Context, _ string, _ int64) {
		atomic.AddInt32(&starts, 1)
		started <- struct{}{}
		<-ctx.Done()
	}

	var messages []string
	app.send = func(ctx context.Context, _ *bot.Bot, _ int64, message string) {
		messages = append(messages, message)
	}

	update := &models.Update{
		Message: &models.Message{
			Text: "/notify_start",
			Chat: models.Chat{ID: 7},
			From: &models.User{ID: 99},
		},
	}

	app.notifyStartHandler(context.Background(), nil, update)
	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for notification routine to start")
	}

	app.notifyStartHandler(context.Background(), nil, update)
	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for second notification routine to start")
	}

	if atomic.LoadInt32(&starts) != 2 {
		t.Fatalf("expected 2 routine starts, got %d", atomic.LoadInt32(&starts))
	}
	if len(messages) < 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0] != "Notifications Active" {
		t.Fatalf("unexpected first message: %q", messages[0])
	}
	if messages[1] != "Notifications Updated" {
		t.Fatalf("unexpected second message: %q", messages[1])
	}

	app.notify.StopAll()
}

func TestBroadcastHandler(t *testing.T) {
	store := &fakeStore{
		notifications: []Notification{
			{UserID: "1", ChatID: 100},
			{UserID: "2", ChatID: 200},
			{UserID: "3", ChatID: 100},
		},
	}
	client := &noopBalanceClient{}
	app := NewApp(store, client, nil, NewNotificationManager(), "99")

	var sent []int64
	var messages []string
	app.send = func(ctx context.Context, _ *bot.Bot, chatID int64, message string) {
		sent = append(sent, chatID)
		messages = append(messages, message)
	}

	update := &models.Update{
		Message: &models.Message{
			Text: "/broadcast hello everyone",
			Chat: models.Chat{ID: 9},
			From: &models.User{ID: 99},
		},
	}

	app.broadcastHandler(context.Background(), nil, update)

	if len(sent) != 3 {
		t.Fatalf("expected 3 messages sent, got %d", len(sent))
	}
	if messages[0] != "hello everyone" || messages[1] != "hello everyone" {
		t.Fatalf("unexpected broadcast content: %v", messages)
	}
	if messages[2] != "Broadcast sent" {
		t.Fatalf("expected confirmation message, got %q", messages[2])
	}
}

type noopBalanceClient struct{}

func (c *noopBalanceClient) GetPoolBalance(poolID string) (int64, error) {
	return 0, nil
}

func (c *noopBalanceClient) GetDelegationBalance(delegationID string) (int64, error) {
	return 0, nil
}
