package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
)

// SendMessage отправляет сообщение через Telegram API.
func SendMessage(sender *tgbotapi.BotAPI, chatID int64, m chat.Message) (err error) {
	if m.IsEmpty() {
		return nil
	}

	switch msgContent := m.Content.(type) {
	case string:
		msg := tgbotapi.NewMessage(chatID, msgContent)
		msg.ParseMode = tgbotapi.ModeHTML

		_, err = sender.Send(msg)

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

	case content.Typing:
		typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)

		_, err = sender.Request(typing)

	default:
		err = ErrTelegramUnsupportedMessageType
	}

	return
}
