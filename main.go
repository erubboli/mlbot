package main

import (
	"context"
	"fmt"
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
	defer db.Close()

	log.Printf("Starting bot with token %s", config.BotToken)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	ctx = context.WithValue(ctx, "db", db)

	defer cancel()
	opts := []bot.Option{
		bot.WithMiddlewares(showMessageWithUserName),
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(config.BotToken, opts...)
	if nil != err {
		log.Panic(err)
	}

	b = registerHandlers(b)
	recoverPastNotifications(ctx, b, db)

	b.Start(ctx)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
	log.Println("Shutting down...")
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
