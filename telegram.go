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

func bold(s string) string              { return "<b>" + s + "</b>" }
func code(s string) string              { return "<code>" + s + "</code>" }
func htmlLink(text, url string) string  { return fmt.Sprintf(`<a href="%s">%s</a>`, url, text) }

func shortID(id string) string {
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "…" + id[len(id)-6:]
}

func poolLink(id string) string {
	return htmlLink(shortID(id), explorerBaseURL+"/pool/"+id)
}

func delegationLink(id string) string {
	return htmlLink(shortID(id), explorerBaseURL+"/delegation/"+id)
}

func escapeHTML(input string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(input)
}

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
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/debug_status", bot.MatchTypeContains, a.debugStatusHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/debug_stop", bot.MatchTypeContains, a.debugStopHandler)
	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/debug_start", bot.MatchTypeContains, a.debugStartHandler)
	//	a.bot.RegisterHandler(bot.HandlerTypeMessageText, "/address_add", bot.MatchTypeContains, a.addressAddHandler)
}

func (a *App) helloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := "ℹ️ " + bold("Mintlayer Bot") + "\n\n"
	msg += bold("⛏️ Pools") + "\n"
	msg += "/pool_add — Add a pool\n"
	msg += "/pool_remove — Remove a pool\n"
	msg += "/pool_list — List your pools\n\n"
	msg += bold("🤝 Delegations") + "\n"
	msg += "/delegation_add — Add a delegation\n"
	msg += "/delegation_remove — Remove a delegation\n"
	msg += "/delegation_list — List your delegations\n\n"
	msg += bold("💰 Balance") + "\n"
	msg += "/balance — Get total balance\n\n"
	msg += bold("🔔 Notifications") + "\n"
	msg += "/notify_start — Subscribe to balance changes\n"
	msg += "/notify_stop — Unsubscribe\n"
	msg += "/notify_status — Check status\n"
	if a.adminUser == fmt.Sprint(update.Message.From.ID) {
		msg += "\n" + bold("🔧 Admin") + "\n"
		msg += "/broadcast — Broadcast to all notification channels\n"
		msg += "/debug_status — Notification status for a user\n"
		msg += "/debug_stop — Stop notifications for a user\n"
		msg += "/debug_start — Start notifications for a user\n"
	}
	a.sendMessage(ctx, b, update.Message.Chat.ID, msg)
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
		a.sendCommandError(ctx, b, chatID)
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
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: /pool_add "+code("poolID"))
		return
	}

	poolID := parts[1]
	if !validateBech32Address(poolID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid pool ID")
		return
	}

	err := a.store.AddPool(ctx, fmt.Sprint(userID), poolID)
	if err != nil {
		log.Printf("Error adding pool: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "✅ Pool added")
	}
}

func (a *App) removePoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: /pool_remove "+code("poolID"))
		return
	}

	poolID := parts[1]
	if !validateBech32Address(poolID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid pool ID")
		return
	}
	err := a.store.RemovePool(ctx, fmt.Sprint(userID), poolID)
	if err != nil {
		log.Printf("Error removing pool: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "🗑️ Pool removed")
	}
}

func (a *App) listPoolHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	p := message.NewPrinter(language.AmericanEnglish)

	pools, err := a.store.GetPools(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing pools: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
	} else {
		if len(pools) == 0 {
			a.sendMessage(ctx, b, update.Message.Chat.ID, "⛏️ You have no pools yet. Add one with /pool_add")
		} else {
			balances, err := runFetchMapWithLimit(pools, 10, func(poolID string) (int64, error) {
				return a.client.GetPoolBalance(poolID)
			})
			if err != nil {
				log.Printf("Error getting pool balance: %v", err)
				a.sendCommandError(ctx, b, update.Message.Chat.ID)
				return
			}
			poolMessage := "⛏️ " + bold(fmt.Sprintf("Your Pools (%d)", len(pools))) + "\n\n"
			for _, poolID := range pools {
				balance := balances[poolID]
				if balance == 0 {
					poolMessage += "• " + poolLink(poolID) + "  💀 Decommissioned\n"
				} else {
					poolMessage += p.Sprintf("• %s  💰 %v ML\n", poolLink(poolID), balance)
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
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: /delegation_add "+code("delegationID"))
		return
	}

	delegationID := parts[1]
	if !validateBech32Address(delegationID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid delegation ID")
		return
	}

	err := a.store.AddDelegation(ctx, fmt.Sprint(userID), delegationID)
	if err != nil {
		log.Printf("Error adding delegation: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "✅ Delegation added")
	}
}

func (a *App) removeDelegationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "Usage: /delegation_remove "+code("delegationID"))
		return
	}

	delegationID := parts[1]

	if !validateBech32Address(delegationID) {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid delegation ID")
		return
	}

	err := a.store.RemoveDelegation(ctx, fmt.Sprint(userID), delegationID)
	if err != nil {
		log.Printf("Error removing delegation: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
	} else {
		a.sendMessage(ctx, b, update.Message.Chat.ID, "🗑️ Delegation removed")
	}
}

func (a *App) listDelegationsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	p := message.NewPrinter(language.AmericanEnglish)

	delegations, err := a.store.GetDelegations(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error listing delegations: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
	} else {
		if len(delegations) == 0 {
			a.sendMessage(ctx, b, update.Message.Chat.ID, "🤝 You have no delegations yet. Add one with /delegation_add")
		} else {
			balances, err := runFetchMapWithLimit(delegations, 10, func(delegationID string) (int64, error) {
				return a.client.GetDelegationBalance(delegationID)
			})
			if err != nil {
				log.Printf("Error getting delegation balance: %v", err)
				a.sendCommandError(ctx, b, update.Message.Chat.ID)
				return
			}
			delegationMessage := "🤝 " + bold(fmt.Sprintf("Your Delegations (%d)", len(delegations))) + "\n\n"
			for _, delegationID := range delegations {
				balance := balances[delegationID]
				delegationMessage += p.Sprintf("• %s  💰 %v ML\n", delegationLink(delegationID), balance)
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
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
		return
	}

	var poolsTotalBalance int64
	var delegationsTotalBalance int64

	delegations, err := a.store.GetDelegations(ctx, fmt.Sprint(userID))
	if err != nil {
		log.Printf("Error getting delegations: %v", err)
		a.sendCommandError(ctx, b, update.Message.Chat.ID)
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
			a.sendCommandError(ctx, b, update.Message.Chat.ID)
			return
		}
	}

	p := message.NewPrinter(language.AmericanEnglish)
	msg := "💰 " + bold("Balance Summary") + "\n\n"
	msg += fmt.Sprintf("⛏️ Pools (%d): %s\n", len(pools), bold(p.Sprintf("%v ML", poolsTotalBalance)))
	msg += fmt.Sprintf("🤝 Delegations (%d): %s\n", len(delegations), bold(p.Sprintf("%v ML", delegationsTotalBalance)))
	msg += "─────────────────\n"
	msg += "📊 Total: " + bold(p.Sprintf("%v ML", poolsTotalBalance+delegationsTotalBalance))

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


func (a *App) notifyStartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if a.notify.Active(userID) {
		chatIDs, err := a.store.GetNotificationChatIDs(ctx, userID)
		if err == nil && len(chatIDs) == 1 && chatIDs[0] == chatID {
			a.sendMessage(ctx, b, chatID, "🔔 Notifications already active")
			return
		}
	}

	if err := a.store.ReplaceNotificationsChannel(ctx, userID, chatID); err != nil {
		log.Printf("Error updating notification channel: %v", err)
		a.sendCommandError(ctx, b, chatID)
		return
	}

	wasActive := a.notify.Active(userID)
	if wasActive {
		a.notify.Stop(userID)
	}

	if !a.notify.Start(a.appCtx, userID, func(ctx context.Context) {
		a.startNotify(ctx, userID, chatID)
	}) {
		a.sendMessage(ctx, b, chatID, "🔔 Notifications already active")
		return
	}

	if wasActive {
		a.sendMessage(ctx, b, chatID, "🔔 Notifications updated")
	} else {
		a.sendMessage(ctx, b, chatID, "🔔 Notifications started")
	}
}

func (a *App) broadcastHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if a.adminUser == "" || userID != a.adminUser {
		a.sendMessage(ctx, b, chatID, "🚫 Unauthorized")
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, chatID, "Usage: /broadcast "+code("message"))
		return
	}
	message := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/broadcast"))
	if message == "" {
		a.sendMessage(ctx, b, chatID, "Usage: /broadcast "+code("message"))
		return
	}
	message = escapeHTML(message)

	notifications, err := a.store.GetAllNotifications(ctx)
	if err != nil {
		log.Printf("Error getting notifications: %v", err)
		a.sendCommandError(ctx, b, chatID)
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
	a.sendMessage(ctx, b, chatID, "✅ Broadcast sent")
}

func (a *App) debugStatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if a.adminUser == "" || userID != a.adminUser {
		a.sendMessage(ctx, b, chatID, "🚫 Unauthorized")
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, chatID, "Usage: /debug_status "+code("user_id"))
		return
	}

	targetUserID := parts[1]
	chatIDs, err := a.store.GetNotificationChatIDs(ctx, targetUserID)
	if err != nil {
		log.Printf("Error getting notification channels: %v", err)
		a.sendCommandError(ctx, b, chatID)
		return
	}

	status := "inactive"
	if a.notify.Active(targetUserID) {
		status = "active"
	}

	msg := fmt.Sprintf("User %s: %s, channels=%d", targetUserID, status, len(chatIDs))
	if len(chatIDs) > 0 {
		msg += fmt.Sprintf(" (%v)", chatIDs)
	}
	a.sendMessage(ctx, b, chatID, msg)
}

func (a *App) debugStopHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if a.adminUser == "" || userID != a.adminUser {
		a.sendMessage(ctx, b, chatID, "🚫 Unauthorized")
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, chatID, "Usage: /debug_stop "+code("user_id"))
		return
	}

	targetUserID := parts[1]
	if a.notify.Stop(targetUserID) {
		a.sendMessage(ctx, b, chatID, "✅ Notifications stopped")
	} else {
		a.sendMessage(ctx, b, chatID, "⚠️ Notifications not active")
	}
}

func (a *App) debugStartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	if a.adminUser == "" || userID != a.adminUser {
		a.sendMessage(ctx, b, chatID, "🚫 Unauthorized")
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		a.sendMessage(ctx, b, chatID, "Usage: /debug_start "+code("user_id")+" [chat_id]")
		return
	}

	targetUserID := parts[1]
	var targetChatID int64

	if len(parts) >= 3 {
		parsed, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			a.sendMessage(ctx, b, chatID, "Usage: /debug_start "+code("user_id")+" [chat_id]")
			return
		}
		targetChatID = parsed
		if err := a.store.ReplaceNotificationsChannel(ctx, targetUserID, targetChatID); err != nil {
			log.Printf("Error updating notification channel: %v", err)
			a.sendCommandError(ctx, b, chatID)
			return
		}
	} else {
		chatIDs, err := a.store.GetNotificationChatIDs(ctx, targetUserID)
		if err != nil {
			log.Printf("Error getting notification channels: %v", err)
			a.sendCommandError(ctx, b, chatID)
			return
		}
		if len(chatIDs) == 0 {
			a.sendMessage(ctx, b, chatID, "⚠️ No notification channels found for user")
			return
		}
		if len(chatIDs) > 1 {
			a.sendMessage(ctx, b, chatID, "⚠️ Multiple channels found; provide chat_id")
			return
		}
		targetChatID = chatIDs[0]
	}

	if a.notify.Active(targetUserID) {
		a.notify.Stop(targetUserID)
	}
	if !a.notify.Start(a.appCtx, targetUserID, func(ctx context.Context) {
		a.startNotify(ctx, targetUserID, targetChatID)
	}) {
		a.sendMessage(ctx, b, chatID, "⚠️ Notifications already active")
		return
	}

	a.sendMessage(ctx, b, chatID, "✅ Notifications started")
}

func (a *App) notifyStatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := fmt.Sprint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	if a.notify.Active(userID) {
		a.sendMessage(ctx, b, chatID, "🔔 Subscribed")
	} else {
		a.sendMessage(ctx, b, chatID, "❌ Not subscribed")
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
			a.sendCommandError(ctx, b, chatID)
			return
		}
		a.sendMessage(ctx, b, chatID, "🔕 Notifications stopped")
	} else {
		log.Println("User not subscribed to notifications: ", userID)
		a.sendMessage(ctx, b, chatID, "❌ Not subscribed")
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
		return
	}

	runTasksWithLimit(delegations, 10, func(delegationID string) {
		/*start := delegationID[:10]
		end := delegationID[len(delegationID)-10:]
		printableDelegationID := start + "..." + end*/

		new_balance, err := a.client.GetDelegationBalance(delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			return
		}
		old_balance, err := a.store.GetDelegationBalance(ctx, userID, delegationID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			return
		}
		if new_balance != old_balance {
			if err = a.store.UpdateDelegationBalance(ctx, userID, delegationID, new_balance); err != nil {
				log.Printf("Error updating balance: %v", err)
				return
			}
			p := message.NewPrinter(language.AmericanEnglish)
			if new_balance >= old_balance {
				diff := new_balance - old_balance
				msg := fmt.Sprintf("📈 Delegation balance increased\n%s: +%s (now %s)",
					delegationLink(delegationID),
					p.Sprintf("%v ML", diff),
					p.Sprintf("%v ML", new_balance))
				a.sendMessage(ctx, a.bot, chatID, msg)
			} else {
				diff := old_balance - new_balance
				msg := fmt.Sprintf("📉 Delegation balance decreased\n%s: −%s (now %s)",
					delegationLink(delegationID),
					p.Sprintf("%v ML", diff),
					p.Sprintf("%v ML", new_balance))
				a.sendMessage(ctx, a.bot, chatID, msg)
			}
		}
	})
}

func (a *App) notifyPoolsBalanceChanges(ctx context.Context, userID string, chatID int64) {
	pools, err := a.store.GetPools(ctx, userID)
	if err != nil {
		log.Printf("Error getting pools: %v", err)
		return
	}

	runTasksWithLimit(pools, 10, func(poolID string) {
		/*start := poolID[:10]
		end := poolID[len(poolID)-10:]
		printablePoolID := start + "..." + end*/

		new_balance, err := a.client.GetPoolBalance(poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			return
		}
		old_balance, err := a.store.GetPoolBalance(ctx, userID, poolID)
		if err != nil {
			log.Printf("Error fetching balance: %v", err)
			return
		}
		if new_balance != old_balance {
			if err = a.store.UpdatePoolBalance(ctx, userID, poolID, new_balance); err != nil {
				log.Printf("Error updating balance: %v", err)
				return
			}
			p := message.NewPrinter(language.AmericanEnglish)
			if new_balance >= old_balance {
				diff := new_balance - old_balance
				msg := fmt.Sprintf("📈 Pool balance increased\n%s: +%s (now %s)",
					poolLink(poolID),
					p.Sprintf("%v ML", diff),
					p.Sprintf("%v ML", new_balance))
				a.sendMessage(ctx, a.bot, chatID, msg)
			} else {
				diff := old_balance - new_balance
				msg := fmt.Sprintf("📉 Pool balance decreased\n%s: −%s (now %s)",
					poolLink(poolID),
					p.Sprintf("%v ML", diff),
					p.Sprintf("%v ML", new_balance))
				a.sendMessage(ctx, a.bot, chatID, msg)
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
		userID, chatID := notification.UserID, notification.ChatID
		log.Printf("Recovering notification for user %v, on chan %v \n", userID, chatID)
		a.notify.Start(ctx, userID, func(ctx context.Context) {
			a.startNotify(ctx, userID, chatID)
		})
	}
}
