package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	_ "github.com/mattn/go-sqlite3"
)

const (
	configFile = "config.json"
)

func main() {
	config, err := readConfig(configFile)
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	db := initDB("pools.db")

	log.Printf("Starting bot with token %s", config.BotToken)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()
	opts := []bot.Option{
		bot.WithMiddlewares(showMessageWithUserName),
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(config.BotToken, opts...)
	if nil != err {
		log.Panic(err)
	}

	store := NewSQLStore(db)
	client := &HTTPBalanceClient{}
	app := NewApp(store, client, b, NewNotificationManager())
	app.registerHandlers()
	app.recoverPastNotifications(ctx)

	botDone := make(chan struct{})
	go func() {
		b.Start(ctx)
		close(botDone)
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down...")
	app.notify.StopAll()
	<-botDone
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

func showMessageWithUserName(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			log.Printf("%s: %s", update.Message.From.FirstName, update.Message.Text)
		}
		next(ctx, b, update)
	}
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		log.Println("Received message:", update.Message.Text)
	}
}
