package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
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
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/broadcast", bot.MatchTypeContains, a.broadcastHandler)
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
	// if current user is admin, show these admin specific commands
	if a.adminUser == fmt.Sprint(update.Message.From.ID) {
		helpMessage += "`/broadcast <message>` : *Admin only: broadcast to notification channels*\n"
	}

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

	err := a.store.AddMonitoredAddress(ctx, fmt.Sprint(userID), address, threshold, notifyOnChange, update.Message.Chat.ID)
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

	err := a.store.AddPool(ctx, fmt.Sprint(userID), poolID)
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
	err := a.store.RemovePool(ctx, fmt.Sprint(userID), poolID)
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

	pools, err := a.store.GetPools(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing pools: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error listing pools: "+err.Error())
	} else {
		if len(pools) == 0 {
			a.sendMessage(ctx, b, update.Message.Chat.ID, "You have no pools")
		} else {
			balances, err := runFetchMapWithLimit(pools, 10, func(poolID string) (int64, error) {
				return a.client.GetPoolBalance(poolID)
			})
			if err != nil {
				log.Printf("Error getting pool balance: %v", err)
				a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pool balance: "+err.Error())
				return
			}
			poolMessage := "Your pools:\n"
			for _, poolID := range pools {
				balance := balances[poolID]
				if balance == 0 {
					poolMessage += p.Sprintf("`%v`: `decommissioned` \n", poolID)
				} else {
					poolMessage += p.Sprintf("`%v`: %v ML \n", poolID, balance)
				}
			}
			a.sendLongMessage(ctx, b, update.Message.Chat.ID, poolMessage)
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

	err := a.store.AddDelegation(ctx, fmt.Sprint(userID), delegationID)
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

	err := a.store.RemoveDelegation(ctx, fmt.Sprint(userID), delegationID)
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

	delegations, err := a.store.GetDelegations(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing delegations: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error listing delegations: "+err.Error())
	} else {
		if len(delegations) == 0 {
			a.sendMessage(ctx, b, update.Message.Chat.ID, "You have no delegations")
		} else {
			balances, err := runFetchMapWithLimit(delegations, 10, func(delegationID string) (int64, error) {
				return a.client.GetDelegationBalance(delegationID)
			})
			if err != nil {
				log.Printf("Error getting delegation balance: %v", err)
				a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegation balance: "+err.Error())
				return
			}
			delegationMessage := "Your delegations:\n"
			for _, delegationID := range delegations {
				balance := balances[delegationID]
				delegationMessage += p.Sprintf("`%v`: %v ML \n", delegationID, balance)
			}
			a.sendLongMessage(ctx, b, update.Message.Chat.ID, delegationMessage)
		}
	}
}

func (a *App) balanceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID

	pools, err := a.store.GetPools(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting pools: "+err.Error())
		return
	}

	var poolsTotalBalance int64
	var delegationsTotalBalance int64

	delegations, err := a.store.GetDelegations(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting delegations: "+err.Error())
		return
	}

	errCh := make(chan error, 2)

	go func() {
		poolErr := runWithLimit(pools, 10, func(poolID string) (int64, error) {
			return a.client.GetPoolBalance(poolID)
		}, func(balance int64) {
			poolsTotalBalance += balance
		})
		errCh <- poolErr
	}()

	go func() {
		delegationErr := runWithLimit(delegations, 10, func(delegationID string) (int64, error) {
			return a.client.GetDelegationBalance(delegationID)
		}, func(balance int64) {
			delegationsTotalBalance += balance
		})
		errCh <- delegationErr
	}()

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			log.Printf("Error getting balance: %v", err)
			a.sendMessage(ctx, b, update.Message.Chat.ID, "Error getting balance: "+err.Error())
			return
		}
	}

	p := message.NewPrinter(language.AmericanEnglish)
	msg := p.Sprintf("`%v` pools: `%v ML`\n", len(pools), poolsTotalBalance)
	msg += p.Sprintf("`%v` delegations: `%v ML`\n", len(delegations), delegationsTotalBalance)
	msg += p.Sprintf("Total: `%v ML`", poolsTotalBalance+delegationsTotalBalance)

	a.sendMessage(ctx, b, update.Message.Chat.ID, msg)
}

func runWithLimit(ids []string, limit int, fetch func(id string) (int64, error), add func(balance int64)) error {
	if len(ids) == 0 {
		return nil
	}
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)
	sem := make(chan struct{}, limit)

	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			balance, err := fetch(id)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			add(balance)
		}(id)
	}

	wg.Wait()
	return firstErr
}

func runTasksWithLimit(ids []string, limit int, task func(id string)) {
	if len(ids) == 0 {
		return
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, limit)

	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			task(id)
		}(id)
	}

	wg.Wait()
}

func runFetchMapWithLimit(ids []string, limit int, fetch func(id string) (int64, error)) (map[string]int64, error) {
	if len(ids) == 0 {
		return map[string]int64{}, nil
	}
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
		results  = make(map[string]int64, len(ids))
	)
	sem := make(chan struct{}, limit)

	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			balance, err := fetch(id)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			results[id] = balance
		}(id)
	}

	wg.Wait()
	return results, firstErr
}

func splitMessage(message string, limit int) []string {
	if len(message) <= limit {
		return []string{message}
	}
	lines := strings.Split(message, "\n")
	var chunks []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			chunks = append(chunks, current.String())
			current.Reset()
		}
	}

	for _, line := range lines {
		lineLen := len(line)
		if lineLen > limit {
			flush()
			for start := 0; start < lineLen; start += limit {
				end := start + limit
				if end > lineLen {
					end = lineLen
				}
				chunks = append(chunks, line[start:end])
			}
			continue
		}

		if current.Len() == 0 {
			current.WriteString(line)
			continue
		}
		if current.Len()+1+lineLen <= limit {
			current.WriteString("\n")
			current.WriteString(line)
		} else {
			flush()
			current.WriteString(line)
		}
	}

	flush()
	return chunks
}

func (a *App) sendLongMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	const maxMessageSize = 3900
	chunks := splitMessage(message, maxMessageSize)
	for _, chunk := range chunks {
		a.sendMessage(ctx, b, chatID, chunk)
	}
}

func escapeMarkdownV2(input string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(input)
}

func (a *App) notifyStartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if err := a.store.ReplaceNotificationsChannel(ctx, userID, chatID); err != nil {
		log.Printf("Error updating notification channel: %v", err)
		a.sendMessage(ctx, b, chatID, "Error updating notification channel: "+err.Error())
		return
	}

	wasActive := a.notify.Active(userID)
	if wasActive {
		a.notify.Stop(userID)
	}

	if !a.notify.Start(a.appCtx, userID, func(ctx context.Context) {
		a.startNotify(ctx, userID, chatID)
	}) {
		a.sendMessage(ctx, b, chatID, "Notification already Active")
		return
	}

	if wasActive {
		a.sendMessage(ctx, b, chatID, "Notifications Updated")
	} else {
		a.sendMessage(ctx, b, chatID, "Notifications Active")
	}
}

func (a *App) broadcastHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if a.adminUser == "" || userID != a.adminUser {
		a.sendMessage(ctx, b, chatID, "Unauthorized")
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, chatID, "Usage: `/broadcast <message>`")
		return
	}
	message := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/broadcast"))
	if message == "" {
		a.sendMessage(ctx, b, chatID, "Usage: `/broadcast <message>`")
		return
	}
	message = escapeMarkdownV2(message)

	notifications, err := a.store.GetAllNotifications(ctx)
	if err != nil {
		log.Printf("Error getting notifications: %v", err)
		a.sendMessage(ctx, b, chatID, "Error fetching notification channels: "+err.Error())
		return
	}

	seen := make(map[int64]struct{})
	for _, notification := range notifications {
		if _, exists := seen[notification.ChatID]; exists {
			continue
		}
		seen[notification.ChatID] = struct{}{}
		a.sendMessage(ctx, b, notification.ChatID, message)
	}
	a.sendMessage(ctx, b, chatID, "Broadcast sent")
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
		err := a.store.RemoveNotification(ctx, userID, chatID)
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
	delegations, err := a.store.GetDelegations(ctx, userID)
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		a.sendMessage(ctx, a.bot, chatID, "Error getting delegations: "+err.Error())
		return
	}

	runTasksWithLimit(delegations, 10, func(delegationID string) {
		/*start := delegationID[:10]
		end := delegationID[len(delegationID)-10:]
		printableDelegationID := start + "..." + end*/

		new_balance, err := a.client.GetDelegationBalance(delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			return
		}
		old_balance, err := a.store.GetDelegationBalance(ctx, userID, delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			return
		}
		if new_balance != old_balance {
			err = a.store.UpdateDelegationBalance(ctx, userID, delegationID, new_balance)

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
	})
}

func (a *App) notifyPoolsBalanceChanges(ctx context.Context, userID string, chatID int64) {
	pools, err := a.store.GetPools(ctx, userID)
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		a.sendMessage(ctx, a.bot, chatID, "Error getting pools: "+err.Error())
		return
	}

	runTasksWithLimit(pools, 10, func(poolID string) {
		/*start := poolID[:10]
		end := poolID[len(poolID)-10:]
		printablePoolID := start + "..." + end*/

		new_balance, err := a.client.GetPoolBalance(poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			return
		}
		old_balance, err := a.store.GetPoolBalance(ctx, userID, poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			a.sendMessage(ctx, a.bot, chatID, "Error fetching balance: "+err.Error())
			return
		}
		if new_balance != old_balance {
			err = a.store.UpdatePoolBalance(ctx, userID, poolID, new_balance)

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
	})
}

func (a *App) recoverPastNotifications(ctx context.Context) {
	notifications, err := a.store.GetAllNotifications(ctx)

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
