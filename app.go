package main

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type App struct {
	store       Store
	client      BalanceClient
	bot         *bot.Bot
	notify      *NotificationManager
	send        func(ctx context.Context, b *bot.Bot, chatID int64, message string)
	startNotify func(ctx context.Context, userID string, chatID int64)
}

func NewApp(store Store, client BalanceClient, b *bot.Bot, notify *NotificationManager) *App {
	app := &App{
		store:  store,
		client: client,
		bot:    b,
		notify: notify,
	}
	app.send = defaultSendMessage
	app.startNotify = app.notifyBalanceChangesRoutine
	return app
}

func (a *App) sendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	if a.send != nil {
		a.send(ctx, b, chatID, message)
		return
	}
	defaultSendMessage(ctx, b, chatID, message)
}

func defaultSendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}
