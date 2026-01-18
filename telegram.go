package main

import (
	"context"
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

func (a *App) registerHandlers() {
	//hendle commands: pool_add
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/hello", bot.MatchTypeContains, a.helloHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeContains, a.helloHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/pool_add", bot.MatchTypeContains, a.addPoolHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/pool_remove", bot.MatchTypeContains, a.removePoolHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/pool_list", bot.MatchTypeContains, a.listPoolHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/delegation_add", bot.MatchTypeContains, a.addDelegationHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/delegation_remove", bot.MatchTypeContains, a.removeDelegationHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/delegation_list", bot.MatchTypeContains, a.listDelegationsHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeContains, a.balanceHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/notify_start", bot.MatchTypeContains, a.notifyStartHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/notify_stop", bot.MatchTypeContains, a.notifyStopHanlder)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/notify_status", bot.MatchTypeContains, a.notifyStatusHandler)
	//	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/address_add", bot.MatchTypeContains, a.addressAddHandler)
}

func (a *App) helloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	a.sendMessage(ctx, b, update.Message.Chat.ID, helpMessage)
}

func (a *App) addressAddHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) == 0 {
		a.sendMessage(ctx, b, chatID, "Usage: `/address_add <address> [threshold]`")
		return
	}

	address := parts[1]
	if !validateBech32Address(address) {
		a.sendMessage(ctx, b, chatID, "Invalid address")
		return
	}
	var threshold int
	var notifyOnChange bool = true

	if len(parts) > 2 {
		var err error
		threshold, err = strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("Invalid threshold provided, defaulting to notify on change: %v", err)
			a.sendMessage(ctx, b, chatID, "Usage: `/address_add <address> [threshold]`")
			return
		} else {
			notifyOnChange = false
		}
	}

	err := a.store.AddMonitoredAddress(fmt.Sprint(userID), address, threshold, notifyOnChange, update.Message.Chat.ID)
	if err != nil {
		log.Printf("Error adding monitored address: %v", err)
		a.sendMessage(ctx, b, chatID, "Error adding address for monitoring: "+err.Error())
		return
	}
	a.sendMessage(ctx, b, chatID, fmt.Sprintf("`%s` added for monitoring, ensure you start notifications with `/notify_start`", address))
}

func (a *App) addPoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("addPoolHandler")
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		log.Printf("no parameters")
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/pool_add <poolID>`")
		return
	}

	poolID := parts[1]
	if !validateBech32Address(poolID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Invalid pool ID")
		return
	}

	err := a.store.AddPool(fmt.Sprint(userID), poolID)
	if err != nil {
		log.Printf("Error adding pool: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error adding pool: "+err.Error())
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Pool added")
	}
}

func (a *App) removePoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/pool_remove <poolID>`")
		return
	}

	poolID := parts[1]
	if !validateBech32Address(poolID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Invalid pool ID")
		return
	}
	err := a.store.RemovePool(fmt.Sprint(userID), poolID)
	if err != nil {
		log.Printf("Error removing pool: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error removing pool: "+err.Error())
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Pool removed")
	}
}

func (a *App) listPoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	p := message.NewPrinter(language.AmericanEnglish)

	pools, err := a.store.GetPools(fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing pools: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error listing pools: "+err.Error())
	} else {
		if len(pools) == 0 {
			a.sendMessage(ctx, b, update.Message.Chat.ID, "You have no pools")
		} else {
			poolMessage := "Your pools:\n"
			for _, poolID := range pools {
				balance, err := a.client.GetPoolBalance(poolID)
				if err != nil {
					log.Printf("Error getting pool balance: %v", err)
					a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pool balance: "+err.Error())
					return
				}
				if balance == 0 {
					poolMessage += p.Sprintf("`%v`: `decommissioned` \n", poolID)
				} else {
					poolMessage += p.Sprintf("`%v`: %v ML \n", poolID, balance)
				}
			}
			a.sendMessage(ctx, b, update.Message.Chat.ID, poolMessage)
		}
	}
}

func (a *App) addDelegationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/delegation_add <delegationID>`")
		return
	}

	delegationID := parts[1]
	if !validateBech32Address(delegationID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Invalid delegation ID")
		return
	}

	err := a.store.AddDelegation(fmt.Sprint(userID), delegationID)
	if err != nil {
		log.Printf("Error adding delegation: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error adding delegation: "+err.Error())
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Delegation added")
	}
}

func (a *App) removeDelegationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: `/delegation_remove <delegationID>`")
		return
	}

	delegationID := parts[1]

	if !validateBech32Address(delegationID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Invalid delegation ID")
		return
	}

	err := a.store.RemoveDelegation(fmt.Sprint(userID), delegationID)
	if err != nil {
		log.Printf("Error removing delegation: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error removing delegation: "+err.Error())
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Delegation removed")
	}
}

func (a *App) listDelegationsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	p := message.NewPrinter(language.AmericanEnglish)

	delegations, err := a.store.GetDelegations(fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing delegations: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error listing delegations: "+err.Error())
	} else {
		if len(delegations) == 0 {
			a.sendMessage(ctx, b, update.Message.Chat.ID, "You have no delegations")
		} else {
			delegationMessage := "Your delegations:\n"
			for _, delegationID := range delegations {
				balance, err := a.client.GetDelegationBalance(delegationID)
				if err != nil {
					log.Printf("Error getting delegation balance: %v", err)
					a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegation balance: "+err.Error())
					return
				}
				delegationMessage += p.Sprintf("`%v`: %v ML \n", delegationID, balance)
			}
			a.sendMessage(ctx, b, update.Message.Chat.ID, delegationMessage)
		}
	}
}

func (a *App) balanceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID

	pools, err := a.store.GetPools(fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pools: "+err.Error())
		return
	}

	var poolsTotalBalance int64
	var delegationsTotalBalance int64

	for _, poolID := range pools {
		balance, err := a.client.GetPoolBalance(poolID)
		if err != nil {
			log.Printf("Error getting pool balance: %v", err)
			a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pool balance: "+err.Error())
			return
		}
		poolsTotalBalance += balance
	}

	delegations, err := a.store.GetDelegations(fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegations: "+err.Error())
		return
	}

	for _, delegationID := range delegations {
		balance, err := a.client.GetDelegationBalance(delegationID)
		if err != nil {
			log.Printf("Error getting delegation balance: %v", err)
			a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegation balance: "+err.Error())
			return
		}
		delegationsTotalBalance += balance
	}

	p := message.NewPrinter(language.AmericanEnglish)
	msg := p.Sprintf("`%v` pools: `%v ML`\n", len(pools), poolsTotalBalance)
	msg += p.Sprintf("`%v` delegations: `%v ML`\n", len(delegations), delegationsTotalBalance)
	msg += p.Sprintf("Total: `%v ML`", poolsTotalBalance+delegationsTotalBalance)

	a.sendMessage(ctx, b, update.Message.Chat.ID, msg)
}

func (a *App) notifyStartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	a.store.AddNotification(userID, chatID)
	if !a.notify.Start(ctx, userID, func(ctx context.Context) {
		a.startNotify(ctx, userID, chatID)
	}) {
		a.sendMessage(ctx, b, chatID, "Notification already Active")
		return
	}

	a.sendMessage(ctx, b, chatID, "Notifications Active")
}

func (a *App) notifyStatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	if a.notify.Active(userID) {
		a.sendMessage(ctx, b, chatID, "Subscribed")
	} else {
		a.sendMessage(ctx, b, chatID, "Not Subscribed")
	}
}

func (a *App) notifyStopHanlder(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	if a.notify.Stop(userID) {
		log.Println("Stopping notification for user ", userID)
		err := a.store.RemoveNotification(userID, chatID)
		if err != nil {
			log.Printf("Error removing notification: %v", err)
			a.sendMessage(ctx, b, chatID, "Error removing notification: "+err.Error())
			return
		}
		a.sendMessage(ctx, b, chatID, "Notifications Stopped")
	} else {
		log.Println("User not subscribed to notifications: ", userID)
		a.sendMessage(ctx, b, chatID, "Not Subscribed")
	}
}

func (a *App) notifyBalanceChangesRoutine(ctx context.Context, userID string, chatID int64) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping notification for user ", userID)
			return
		default:
		}
		a.notifyPoolsBalanceChanges(ctx, userID, chatID)
		a.notifyDelegationsBalanceChanges(ctx, userID, chatID)

		select {
		case <-ctx.Done():
			log.Println("Stopping notification for user ", userID)
			return
		case <-ticker.C:
		}
	}
}

func (a *App) notifyDelegationsBalanceChanges(ctx context.Context, userID string, chatID int64) {
	delegations, err := a.store.GetDelegations(userID)
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		a.sendMessage(ctx, a.bot, chatID, "Error getting delegations: "+err.Error())
		return
	}

	for _, delegationID := range delegations {
		/*start := delegationID[:10]
		end := delegationID[len(delegationID)-10:]
		printableDelegationID := start + "..." + end*/

		new_balance, err := a.client.GetDelegationBalance(delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		old_balance, err := a.store.GetDelegationBalance(userID, delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		if new_balance != old_balance {
			err = a.store.UpdateDelegationBalance(userID, delegationID, new_balance)

			p := message.NewPrinter(language.AmericanEnglish)
			if new_balance >= old_balance {
				a.sendMessage(ctx, a.bot, chatID, p.Sprintf("`%s`: \\+%v ML", delegationID, new_balance-old_balance))
			} else {
				a.sendMessage(ctx, a.bot, chatID, p.Sprintf("`%s`: \\-%v ML", delegationID, new_balance-old_balance))
			}
			if err != nil {
				log.Printf("Error updating balance: %v", err)
				a.sendMessage(ctx, a.bot, chatID, "Error updating balance: "+err.Error())
				return
			}
		}
	}
}

func (a *App) notifyPoolsBalanceChanges(ctx context.Context, userID string, chatID int64) {
	pools, err := a.store.GetPools(userID)
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		a.sendMessage(ctx, a.bot, chatID, "Error getting pools: "+err.Error())
		return
	}

	for _, poolID := range pools {
		/*start := poolID[:10]
		end := poolID[len(poolID)-10:]
		printablePoolID := start + "..." + end*/

		new_balance, err := a.client.GetPoolBalance(poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		old_balance, err := a.store.GetPoolBalance(userID, poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			continue
		}
		if new_balance != old_balance {
			err = a.store.UpdatePoolBalance(userID, poolID, new_balance)

			p := message.NewPrinter(language.AmericanEnglish)
			if new_balance >= old_balance {
				a.sendMessage(ctx, a.bot, chatID, p.Sprintf("`%s`: \\+%v ML", poolID, new_balance-old_balance))
			} else {
				a.sendMessage(ctx, a.bot, chatID, p.Sprintf("`%s`: \\-%v ML", poolID, new_balance-old_balance))
			}

			if err != nil {
				log.Printf("Error updating balance: %v", err)
				a.sendMessage(ctx, a.bot, chatID, "Error updating balance: "+err.Error())
				return
			}
		}
	}
}

func (a *App) recoverPastNotifications(ctx context.Context) {
	notifications, err := a.store.GetAllNotifications()

	if err != nil {
		log.Printf("Error getting notifications: %v", err)
		return
	}
	log.Println("Recovering notifications, total: ", len(notifications))

	for _, notification := range notifications {
		log.Printf("Recovering notification for user %v, on chan %v \n", notification.UserID, notification.ChatID)
		a.notify.Start(ctx, notification.UserID, func(ctx context.Context) {
			a.startNotify(ctx, notification.UserID, notification.ChatID)
		})
	}
}
