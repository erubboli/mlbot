package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var cancelNotifyMap = make(map[string]context.CancelFunc)

func sendTgMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

func handleTgCommands(bot *tgbotapi.BotAPI, db *sql.DB) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		userID := fmt.Sprintf("%d", update.Message.From.ID)
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		switch cmd {
		case "start", "help":
			sendHelpMessage(bot, update.Message.Chat.ID)
		case "pool_add":
			handleAddPool(db, userID, args, bot, update.Message.Chat.ID)
		case "pool_remove":
			handleRemovePool(db, userID, args, bot, update.Message.Chat.ID)
		case "pool_list":
			handleListPools(db, userID, bot, update.Message.Chat.ID)
		case "balance":
			handleBalance(db, userID, bot, update.Message.Chat.ID)
		case "notify_start":
			handleNotifyBalanceChange(db, userID, bot, update.Message.Chat.ID)
		case "notify_stop":
			handleStopNotify(userID, bot, update.Message.Chat.ID)
		}
	}
}

func handleAddPool(db *sql.DB, userID, poolID string, bot *tgbotapi.BotAPI, chatID int64) {
	if poolID == "" {
		return
	}
	err := addPool(db, userID, poolID)
	if err != nil {
		log.Printf("Error adding pool: %v", err)
		sendTgMessage(bot, chatID, "Error adding pool: "+err.Error())
	} else {
		sendTgMessage(bot, chatID, "Pool added")
	}
}

func handleRemovePool(db *sql.DB, userID, poolID string, bot *tgbotapi.BotAPI, chatID int64) {
	if poolID == "" {
		return
	}
	err := removePool(db, userID, poolID)
	if err != nil {
		log.Printf("Error removing pool: %v", err)
		sendTgMessage(bot, chatID, "Error removing pool: "+err.Error())
	} else {
		sendTgMessage(bot, chatID, "Pool removed")
	}
}

func handleListPools(db *sql.DB, userID string, bot *tgbotapi.BotAPI, chatID int64) {
	pools, err := getPools(db, userID)
	if err != nil {
		log.Printf("Error listing pools: %v", err)
		sendTgMessage(bot, chatID, "Error listing pools: "+err.Error())
	} else {
		if len(pools) == 0 {
			sendTgMessage(bot, chatID, "You have no pools")
		} else {
			poolMessage := "Your pools:\n"
			for _, pool := range pools {
				poolMessage += "  " + pool + "\n"
			}
			sendTgMessage(bot, chatID, poolMessage)
		}
	}
}

func handleBalance(db *sql.DB, userID string, bot *tgbotapi.BotAPI, chatID int64) {
	pools, err := getPools(db, userID)
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		sendTgMessage(bot, chatID, "Error getting pools: "+err.Error())
		return
	}

	var totalBalance int64
	for _, poolID := range pools {
		balance, err := getPoolBalance(poolID)
		if err != nil {
			log.Printf("Error getting pool balance: %v", err)
			sendTgMessage(bot, chatID, "Error getting pool balance: "+err.Error())
			return
		}
		totalBalance += balance
	}
	p := message.NewPrinter(language.AmericanEnglish)
	msg := p.Sprintf("Total balance of your pools: %v ML", totalBalance)

	sendTgMessage(bot, chatID, msg)
}

func sendHelpMessage(bot *tgbotapi.BotAPI, chatID int64) {
	helpMessage := "Available commands:\n" +
		"/pool_add <poolID> - Add a pool\n" +
		"/pool_remove <poolID> - Remove a pool\n" +
		"/pool_list - List your pools\n" +
		"/balance - Get the total balance of your pools\n" +
		"/notify_start - Notify on balance change\n" +
		"/notify_stop - Stop balance change notifications"
	sendTgMessage(bot, chatID, helpMessage)
}

func notifyBalanceChanges(db *sql.DB, userID string, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping notification for user ", userID)
			return
		default:
			pools, err := getPools(db, userID)
			if err != nil {
				log.Printf("Error getting pools: %v", err)
				sendTgMessage(bot, chatID, "Error getting pools: "+err.Error())
				return
			}
			for _, poolID := range pools {
				new_balance, err := getPoolBalance(poolID)
				if err != nil {
					log.Printf("Error fetching balance: %v", err)
					//sendTgMessage(bot, chatID, "Error fetching balance: "+err.Error())
					continue
				}
				old_balance, err := getPoolBalanceFromDb(db, userID, poolID)
				if err != nil {
					log.Printf("Error fetching balance: %v", err)
					//sendTgMessage(bot, chatID, "Error fetching balance: "+err.Error())
					continue
				}
				if new_balance != old_balance {
					err = updatePoolBalance(db, userID, poolID, new_balance)

					p := message.NewPrinter(language.AmericanEnglish)
					msg := p.Sprintf("Pool %s balance changed: %v ML", poolID, new_balance-old_balance)
					sendTgMessage(bot, chatID, msg)
					if err != nil {
						log.Printf("Error updating balance: %v", err)
						sendTgMessage(bot, chatID, "Error updating balance: "+err.Error())
						return
					}
				}
			}
			time.Sleep(10 * time.Second)
		}
	}

}
func handleNotifyBalanceChange(db *sql.DB, userID string, bot *tgbotapi.BotAPI, chatID int64) {
	if _, exists := cancelNotifyMap[userID]; exists {
		sendTgMessage(bot, chatID, "You're already subscribed to balance change notifications.")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	go notifyBalanceChanges(db, userID, bot, chatID, ctx)
	cancelNotifyMap[userID] = cancel

	sendTgMessage(bot, chatID, "You will now receive notifications for balance changes.")
}

func handleStopNotify(userID string, bot *tgbotapi.BotAPI, chatID int64) {
	if cancelFunc, exists := cancelNotifyMap[userID]; exists {
		log.Println("Stopping notification for user ", userID)
		cancelFunc()
		delete(cancelNotifyMap, userID)
		sendTgMessage(bot, chatID, "Balance change notifications stopped.")
	} else {
		log.Println("User not subscribed to notifications: ", userID)
		sendTgMessage(bot, chatID, "You're not subscribed to balance change notifications.")
	}
}
