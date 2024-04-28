package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var cancelNotifyMap = make(map[string]context.CancelFunc)

func registerHandlers(b *bot.Bot) *bot.Bot {
	//hendle commands: pool_add
	b.RegisterHandler(bot.HandlerTypeMessageText, "/hello", bot.MatchTypeContains, helloHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeContains, helloHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/pool_add", bot.MatchTypeContains, addPoolHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/pool_remove", bot.MatchTypeContains, removePoolHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/pool_list", bot.MatchTypeContains, listPoolHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/delegation_add", bot.MatchTypeContains, addDelegationHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/delegation_remove", bot.MatchTypeContains, removeDelegationHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/delegation_list", bot.MatchTypeContains, listDelegationsHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeContains, balanceHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/notify_start", bot.MatchTypeContains, notifyStartHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/notify_stop", bot.MatchTypeContains, notifyStopHanlder)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/notify_status", bot.MatchTypeContains, notifyStatusHandler)
	//	b.RegisterHandler(bot.HandlerTypeMessageText, "/address_add", bot.MatchTypeContains, addressAddHandler)
	return b
}

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}

func helloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	helpMessage := "Available commands:\n"
	helpMessage += "`/help` : *this help message* \n"
	//	helpMessage += "`/address_add <address> [threshold] ` : *Add a new address for monitoring*\n"
	helpMessage += "`/pool_add <poolID> ` : *Add a pool*\n"
	helpMessage += "`/pool_remove <poolID> ` : *Remove a pool*\n"
	helpMessage += "`/pool_list ` : *List your pools*\n"
	helpMessage += "`/delegation_add <delegationID> ` : *Add a delegation*\n"
	helpMessage += "`/delegation_remove <delegationID> ` : *Remove a delegation*\n"
	helpMessage += "`/delegation_list ` : *List your delegations*\n"
	helpMessage += "`/balance ` : *Get the total balance of your pools*\n"
	helpMessage += "`/notify_start ` : *Notify on balance change*\n"
	helpMessage += "`/notify_stop ` : *Stop balance change notifications*\n"
	helpMessage += "`/notify_status ` : *Check if you're subscribed to balance change notifications*\n"

	sendMessage(ctx, b, update.Message.Chat.ID, helpMessage)
}

func addressAddHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) == 0 {
		sendMessage(ctx, b, chatID, "Usage: `/address_add <address> [threshold]`")
		return
	}

	address := parts[1]
	if !validateBech32Address(address) {
		sendMessage(ctx, b, chatID, "Invalid address")
		return
	}
	var threshold int
	var notifyOnChange bool = true

	if len(parts) > 2 {
		var err error
		threshold, err = strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("Invalid threshold provided, defaulting to notify on change: %v", err)
			sendMessage(ctx, b, chatID, "Usage: `/address_add <address> [threshold]`")
			return
		} else {
			notifyOnChange = false
		}
	}

	err := addMonitoredAddress(db, fmt.Sprint(userID), address, threshold, notifyOnChange, update.Message.Chat.ID)
	if err != nil {
		log.Printf("Error adding monitored address: %v", err)
		sendMessage(ctx, b, chatID, "Error adding address for monitoring: "+err.Error())
		return
	}
	sendMessage(ctx, b, chatID, fmt.Sprintf("`%s` added for monitoring, ensure you start notifications with `/notify_start`", address))
}

func addPoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("addPoolHandler")
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		log.Printf("no parameters")
		sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/pool_add <poolID>`")
		return
	}

	poolID := parts[1]
	if !validateBech32Address(poolID) {
		sendMessage(ctx, b, update.Message.Chat.ID, "Invalid pool ID")
		return
	}

	err := addPool(db, fmt.Sprint(userID), poolID)
	if err != nil {
		log.Printf("Error adding pool: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error adding pool: "+err.Error())
	} else {
		sendMessage(ctx, b, update.Message.Chat.ID, "Pool added")
	}
}

func removePoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/pool_remove <poolID>`")
		return
	}

	poolID := parts[1]
	if !validateBech32Address(poolID) {
		sendMessage(ctx, b, update.Message.Chat.ID, "Invalid pool ID")
		return
	}
	err := removePool(db, fmt.Sprint(userID), poolID)
	if err != nil {
		log.Printf("Error removing pool: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error removing pool: "+err.Error())
	} else {
		sendMessage(ctx, b, update.Message.Chat.ID, "Pool removed")
	}
}

func listPoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	p := message.NewPrinter(language.AmericanEnglish)

	pools, err := getPools(db, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing pools: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error listing pools: "+err.Error())
	} else {
		if len(pools) == 0 {
			sendMessage(ctx, b, update.Message.Chat.ID, "You have no pools")
		} else {
			poolMessage := "Your pools:\n"
			for _, poolID := range pools {
				balance, err := getPoolBalance(poolID)
				if err != nil {
					log.Printf("Error getting pool balance: %v", err)
					sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pool balance: "+err.Error())
					return
				}
				if balance == 0 {
					poolMessage += p.Sprintf("`%v`: `decommissioned` \n", poolID)
				} else {
					poolMessage += p.Sprintf("`%v`: %v ML \n", poolID, balance)
				}
			}
			sendMessage(ctx, b, update.Message.Chat.ID, poolMessage)
		}
	}
}

func addDelegationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/delegation_add <delegationID>`")
		return
	}

	delegationID := parts[1]
	if !validateBech32Address(delegationID) {
		sendMessage(ctx, b, update.Message.Chat.ID, "Invalid delegation ID")
		return
	}

	err := addDelegation(db, fmt.Sprint(userID), delegationID)
	if err != nil {
		log.Printf("Error adding delegation: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error adding delegation: "+err.Error())
	} else {
		sendMessage(ctx, b, update.Message.Chat.ID, "Delegation added")
	}
}

func removeDelegationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/delegation_remove <delegationID>`")
		return
	}

	delegationID := parts[1]

	if !validateBech32Address(delegationID) {
		sendMessage(ctx, b, update.Message.Chat.ID, "Invalid delegation ID")
		return
	}

	err := removeDelegation(db, fmt.Sprint(userID), delegationID)
	if err != nil {
		log.Printf("Error removing delegation: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error removing delegation: "+err.Error())
	} else {
		sendMessage(ctx, b, update.Message.Chat.ID, "Delegation removed")
	}
}

func listDelegationsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID
	p := message.NewPrinter(language.AmericanEnglish)

	delegations, err := getDelegations(db, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing delegations: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error listing delegations: "+err.Error())
	} else {
		if len(delegations) == 0 {
			sendMessage(ctx, b, update.Message.Chat.ID, "You have no delegations")
		} else {
			delegationMessage := "Your delegations:\n"
			for _, delegationID := range delegations {
				balance, err := getDelegationBalance(delegationID)
				if err != nil {
					log.Printf("Error getting delegation balance: %v", err)
					sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegation balance: "+err.Error())
					return
				}
				delegationMessage += p.Sprintf("`%v`: %v ML \n", delegationID, balance)
			}
			sendMessage(ctx, b, update.Message.Chat.ID, delegationMessage)
		}
	}
}

func balanceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := update.Message.From.ID

	pools, err := getPools(db, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pools: "+err.Error())
		return
	}

	var poolsTotalBalance int64
	var delegationsTotalBalance int64

	for _, poolID := range pools {
		balance, err := getPoolBalance(poolID)
		if err != nil {
			log.Printf("Error getting pool balance: %v", err)
			sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pool balance: "+err.Error())
			return
		}
		poolsTotalBalance += balance
	}

	delegations, err := getDelegations(db, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegations: "+err.Error())
		return
	}

	for _, delegationID := range delegations {
		balance, err := getDelegationBalance(delegationID)
		if err != nil {
			log.Printf("Error getting delegation balance: %v", err)
			sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegation balance: "+err.Error())
			return
		}
		delegationsTotalBalance += balance
	}

	p := message.NewPrinter(language.AmericanEnglish)
	msg := p.Sprintf("`%v` pools: `%v ML`\n", len(pools), poolsTotalBalance)
	msg += p.Sprintf("`%v` delegations: `%v ML`\n", len(delegations), delegationsTotalBalance)
	msg += p.Sprintf("Total: `%v ML`", poolsTotalBalance+delegationsTotalBalance)

	sendMessage(ctx, b, update.Message.Chat.ID, msg)
}

func notifyStartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	db := ctx.Value("db").(*sql.DB)
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	addNotification(db, userID, chatID)
	if _, exists := cancelNotifyMap[userID]; exists {
		sendMessage(ctx, b, chatID, "Notification already Active")
		return
	}
	ctx_in, cancel := context.WithCancel(context.Background())
	go notifyBalanceChangesRoutine(ctx_in, db, b, userID, chatID)
	cancelNotifyMap[userID] = cancel

	sendMessage(ctx, b, chatID, "Notifications Active")
}

func notifyStatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	if _, exists := cancelNotifyMap[userID]; exists {
		sendMessage(ctx, b, chatID, "Subscribed")
	} else {
		sendMessage(ctx, b, chatID, "Not Subscribed")
	}
}

func notifyStopHanlder(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	db := ctx.Value("db").(*sql.DB)
	if cancelFunc, exists := cancelNotifyMap[userID]; exists {
		log.Println("Stopping notification for user ", userID)
		cancelFunc()
		delete(cancelNotifyMap, userID)
		err := removeNotification(db, userID, chatID)
		if err != nil {
			log.Printf("Error removing notification: %v", err)
			sendMessage(ctx, b, chatID, "Error removing notification: "+err.Error())
			return
		}
		sendMessage(ctx, b, chatID, "Notifications Stopped")
	} else {
		log.Println("User not subscribed to notifications: ", userID)
		sendMessage(ctx, b, chatID, "Not Subscribed")
	}
}

func notifyBalanceChangesRoutine(ctx context.Context, db *sql.DB, b *bot.Bot, userID string, chatID int64) {
	ctx = context.WithValue(ctx, "db", db)
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping notification for user ", userID)
			return
		default:
			go notifyPoolsBalanceChanges(ctx, b, userID, chatID)
			go notifyDelegationsBalanceChanges(ctx, b, userID, chatID)
			time.Sleep(10 * time.Minute)
		}
	}
}

func notifyDelegationsBalanceChanges(ctx context.Context, b *bot.Bot, userID string, chatID int64) {
	db := ctx.Value("db").(*sql.DB)
	delegations, err := getDelegations(db, userID)
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		sendMessage(ctx, b, chatID, "Error getting delegations: "+err.Error())
		return
	}

	for _, delegationID := range delegations {
		/*start := delegationID[:10]
		end := delegationID[len(delegationID)-10:]
		printableDelegationID := start + "..." + end*/

		new_balance, err := getDelegationBalance(delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			sendMessage(ctx, b, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		old_balance, err := getDelegationBalanceFromDb(db, userID, delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			sendMessage(ctx, b, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		if new_balance != old_balance {
			err = updateDelegationBalance(db, userID, delegationID, new_balance)

			p := message.NewPrinter(language.AmericanEnglish)
			if new_balance >= old_balance {
				sendMessage(ctx, b, chatID, p.Sprintf("`%s`: \\+%v ML", delegationID, new_balance-old_balance))
			} else {
				sendMessage(ctx, b, chatID, p.Sprintf("`%s`: \\-%v ML", delegationID, new_balance-old_balance))
			}
			if err != nil {
				log.Printf("Error updating balance: %v", err)
				sendMessage(ctx, b, chatID, "Error updating balance: "+err.Error())
				return
			}
		}
	}
}

func notifyPoolsBalanceChanges(ctx context.Context, b *bot.Bot, userID string, chatID int64) {
	db := ctx.Value("db").(*sql.DB)
	pools, err := getPools(db, userID)
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		sendMessage(ctx, b, chatID, "Error getting pools: "+err.Error())
		return
	}

	for _, poolID := range pools {
		/*start := poolID[:10]
		end := poolID[len(poolID)-10:]
		printablePoolID := start + "..." + end*/

		new_balance, err := getPoolBalance(poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			sendMessage(ctx, b, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		old_balance, err := getPoolBalanceFromDb(db, userID, poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			sendMessage(ctx, b, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		if new_balance != old_balance {
			err = updatePoolBalance(db, userID, poolID, new_balance)

			p := message.NewPrinter(language.AmericanEnglish)
			if new_balance >= old_balance {
				sendMessage(ctx, b, chatID, p.Sprintf("`%s`: \\+%v ML", poolID, new_balance-old_balance))
			} else {
				sendMessage(ctx, b, chatID, p.Sprintf("`%s`: \\-%v ML", poolID, new_balance-old_balance))
			}

			if err != nil {
				log.Printf("Error updating balance: %v", err)
				sendMessage(ctx, b, chatID, "Error updating balance: "+err.Error())
				return
			}
		}
	}
}

func recoverPastNotifications(ctx context.Context, b *bot.Bot, db *sql.DB) {
	notifications, err := getAllNotifications(db)

	if err != nil {
		log.Printf("Error getting notifications: %v", err)
		return
	}
	log.Println("Recovering notifications, total: ", len(notifications))

	for _, notification := range notifications {
		log.Printf("Recovering notification for user %v, on chan %v \n", notification.UserID, notification.ChatID)
		ctx_in, cancel := context.WithCancel(context.Background())
		go notifyBalanceChangesRoutine(ctx_in, db, b, notification.UserID, notification.ChatID)
		cancelNotifyMap[notification.UserID] = cancel
	}
}
