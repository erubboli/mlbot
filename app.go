package main

import (
	"context"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type App struct {
	store       Store
	client      BalanceClient
	bot         *bot.Bot
	notify      *NotificationManager
	adminUser   string
	appCtx      context.Context
	send        func(ctx context.Context, b *bot.Bot, chatID int64, message string) error
	startNotify func(ctx context.Context, userID string, chatID int64)
}

func NewApp(store Store, client BalanceClient, b *bot.Bot, notify *NotificationManager, adminUser string, appCtx context.Context) *App {
	if appCtx == nil {
		appCtx = context.Background()
	}
	app := &App{
		store:  store,
		client: client,
		bot:    b,
		notify: notify,
		adminUser: adminUser,
		appCtx: appCtx,
	}
	app.send = defaultSendMessage
	app.startNotify = app.notifyBalanceChangesRoutine
	return app
}

func (a *App) sendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	if a.send != nil {
		if err := a.send(ctx, b, chatID, message); err != nil {
			a.handleSendError(ctx, chatID, err)
		}
		return
	}
	if err := defaultSendMessage(ctx, b, chatID, message); err != nil {
		a.handleSendError(ctx, chatID, err)
	}
}

func (a *App) sendCommandError(ctx context.Context, b *bot.Bot, chatID int64) {
	a.sendMessage(ctx, b, chatID, "Something went wrong. Please try again later.")
}

func defaultSendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		if isParseModeError(err) {
			_, retryErr := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   message,
			})
			if retryErr == nil {
				return nil
			}
			log.Println("Error sending message: ", retryErr)
			return retryErr
		}
		log.Println("Error sending message: ", err)
		return err
	}
	return nil
}

func (a *App) handleSendError(ctx context.Context, chatID int64, err error) {
	if a.store == nil {
		return
	}
	if !isChatUnreachableError(err) {
		return
	}
	if cleanupErr := a.store.RemoveNotificationsByChatID(ctx, chatID); cleanupErr != nil {
		log.Printf("Error removing notification for chat %d: %v", chatID, cleanupErr)
	}
}

func isChatUnreachableError(err error) bool {
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "chat not found") ||
		strings.Contains(errText, "bot was blocked by the user")
}

func isParseModeError(err error) bool {
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "can't parse entities")
}
