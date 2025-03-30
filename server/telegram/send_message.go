package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/metrics"
)

// SendMessage sends a message via Telegram API
func SendMessage(sender *tgbotapi.BotAPI, chatID int64, m chat.Message) (err error) {
	if m.IsEmpty() {
		return nil
	}

	switch msgContent := m.Content.(type) {
	case string:
		msg := tgbotapi.NewMessage(chatID, msgContent)
		msg.ParseMode = tgbotapi.ModeHTML

		_, err = sender.Send(msg)
		if err == nil {
			// Increment sent text messages counter
			metrics.TelegramMessagesTotal.WithLabelValues("sent_text").Inc()
		} else {
			metrics.RecordTelegramError("send_text")
		}

	case content.Select:
		var buttons = make([][]tgbotapi.InlineKeyboardButton, 0, len(msgContent.Items))
		for _, item := range msgContent.Items {
			buttons = append(
				buttons,
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(item.Caption, item.Data),
				),
			)
		}

		msg := tgbotapi.NewMessage(chatID, msgContent.Header)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)

		_, err = sender.Send(msg)
		if err == nil {
			// Increment sent selection messages counter
			metrics.TelegramMessagesTotal.WithLabelValues("sent_select").Inc()
		} else {
			metrics.RecordTelegramError("send_select")
		}

	case content.Commands:
		commands := make([]tgbotapi.BotCommand, 0, len(msgContent.Items))
		for _, item := range msgContent.Items {
			commands = append(commands, tgbotapi.BotCommand{
				Command:     item.Name,
				Description: item.Description,
			})
		}

		cmd := tgbotapi.NewSetMyCommands(commands...)

		_, err = sender.Request(cmd)
		if err == nil {
			// Increment commands update counter
			metrics.TelegramMessagesTotal.WithLabelValues("sent_commands").Inc()
		} else {
			metrics.RecordTelegramError("send_commands")
		}

	case content.Typing:
		typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)

		_, err = sender.Request(typing)
		if err == nil {
			// Increment typing action counter
			metrics.TelegramMessagesTotal.WithLabelValues("sent_typing").Inc()
		} else {
			metrics.RecordTelegramError("send_typing")
		}

	default:
		err = ErrTelegramUnsupportedMessageType
		metrics.RecordTelegramError("unsupported_message_type")
	}

	return
}
