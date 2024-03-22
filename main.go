package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	log.Printf("Starting bot with token %s", config.BotToken)
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	db := initDB("pools.db")
	defer db.Close()
	recoverPastNotifications(bot, db)
	go handleTgCommands(bot, db)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c
	log.Println("Shutting down...")
}
