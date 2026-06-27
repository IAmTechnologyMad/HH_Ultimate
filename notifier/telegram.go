package notifier

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func New(token string, chatID int64) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	log.Printf("[Telegram] Authorized as @%s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:    bot,
		chatID: chatID,
	}, nil
}

func (t *TelegramNotifier) SendAlert(message string) error {
	msg := tgbotapi.NewMessage(t.chatID, message)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = false

	_, err := t.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}

	log.Printf("[Telegram] Alert sent successfully")
	return nil
}

func (t *TelegramNotifier) SendStartup(productCount int) error {
	msg := fmt.Sprintf(
		"🤖 <b>Hot Wheels Monitor Started!</b>\n\n"+
			"👀 Watching: FirstCry Hot Wheels\n"+
			"📦 Currently tracking: <b>%d products</b>\n"+
			"⚡ You'll be notified the instant a new or restocked item appears!\n\n"+
			"🏎️ Ready to snag those Hot Wheels!",
		productCount,
	)
	return t.SendAlert(msg)
}

func (t *TelegramNotifier) SendError(errMsg string) error {
	msg := fmt.Sprintf("⚠️ <b>Monitor Error</b>\n\n%s\n\nBot will retry automatically.", errMsg)
	return t.SendAlert(msg)
}
