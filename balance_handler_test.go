package main

import (
	"context"
	"errors"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type fakeStore struct {
	pools         []string
	delegations   []string
	notifications []Notification
}

func (f *fakeStore) AddMonitoredAddress(ctx context.Context, userID, address string, threshold int, notifyOnChange bool, chatID int64) error {
	return nil
}

func (f *fakeStore) AddPool(ctx context.Context, userID, poolID string) error    { return nil }
func (f *fakeStore) RemovePool(ctx context.Context, userID, poolID string) error { return nil }
func (f *fakeStore) GetPools(ctx context.Context, userID string) ([]string, error) {
	return f.pools, nil
}
func (f *fakeStore) AddDelegation(ctx context.Context, userID, delegationID string) error { return nil }
func (f *fakeStore) RemoveDelegation(ctx context.Context, userID, delegationID string) error {
	return nil
}
func (f *fakeStore) GetDelegations(ctx context.Context, userID string) ([]string, error) {
	return f.delegations, nil
}
func (f *fakeStore) GetPoolBalance(ctx context.Context, userID, poolID string) (int64, error) {
	return 0, nil
}
func (f *fakeStore) UpdatePoolBalance(ctx context.Context, userID, poolID string, balance int64) error {
	return nil
}
func (f *fakeStore) GetDelegationBalance(ctx context.Context, userID, delegationID string) (int64, error) {
	return 0, nil
}
func (f *fakeStore) UpdateDelegationBalance(ctx context.Context, userID, delegationID string, balance int64) error {
	return nil
}
func (f *fakeStore) AddNotification(ctx context.Context, userID string, chatID int64) error {
	return nil
}
func (f *fakeStore) RemoveNotification(ctx context.Context, userID string, chatID int64) error {
	return nil
}
func (f *fakeStore) ReplaceNotificationsChannel(ctx context.Context, userID string, chatID int64) error {
	return nil
}
func (f *fakeStore) GetAllNotifications(ctx context.Context) ([]Notification, error) {
	return f.notifications, nil
}

type fakeBalanceClient struct {
	poolBalances       map[string]int64
	delegationBalances map[string]int64
}

func (f *fakeBalanceClient) GetPoolBalance(poolID string) (int64, error) {
	bal, ok := f.poolBalances[poolID]
	if !ok {
		return 0, errors.New("missing pool balance")
	}
	return bal, nil
}

func (f *fakeBalanceClient) GetDelegationBalance(delegationID string) (int64, error) {
	bal, ok := f.delegationBalances[delegationID]
	if !ok {
		return 0, errors.New("missing delegation balance")
	}
	return bal, nil
}

func TestBalanceHandlerAggregatesBalances(t *testing.T) {
	store := &fakeStore{
		pools:       []string{"p1", "p2"},
		delegations: []string{"d1", "d2"},
	}
	client := &fakeBalanceClient{
		poolBalances: map[string]int64{
			"p1": 2,
			"p2": 3,
		},
		delegationBalances: map[string]int64{
			"d1": 1,
			"d2": 2,
		},
	}
	app := NewApp(store, client, nil, NewNotificationManager(), "")

	var lastMessage string
	app.send = func(ctx context.Context, _ *bot.Bot, _ int64, message string) {
		lastMessage = message
	}

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 5},
			From: &models.User{ID: 7},
		},
	}

	app.balanceHandler(context.Background(), nil, update)
	expected := "`2` pools: `5 ML`\n`2` delegations: `3 ML`\nTotal: `8 ML`"
	if lastMessage != expected {
		t.Fatalf("unexpected message:\nexpected: %q\ngot:      %q", expected, lastMessage)
	}
}
