package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/chat/user"
	"github.com/muzykantov/health-gpt/llm"
	"github.com/muzykantov/health-gpt/server"
)

// Определяем основные ошибки при работе с Telegram API.
var (
	ErrTelegramTokenNotProvided       = errors.New("telegram token not provided")
	ErrTelegramLLMNotProvided         = errors.New("telegram llm not provided")
	ErrTelegramUnsupportedMessageType = errors.New("telegram unsupported message type")
	ErrTelegramInvalidMessageContent  = errors.New("telegram invalid message content")
)

// Server управляет взаимодействием с Server Bot API.
type Server struct {
	Token               string
	Handler             server.Handler
	Completion          server.ChatCompleter
	History             server.ChatHistoryReadWriter
	User                server.UserStorage
	Debug               bool
	UnsupportedResponse func() chat.Message
	ErrorLog            *log.Logger
}

// ListenAndServe запускает основной цикл обработки сообщений.
func (t *Server) ListenAndServe(ctx context.Context) error {
	if t.Token == "" {
		return ErrTelegramTokenNotProvided
	}

	chatCompletion := t.Completion
	if chatCompletion == nil {
		chatCompletion = &llm.Mock{
			CompleteChatFn: func(
				ctx context.Context,
				msgs []chat.Message,
			) (chat.Message, error) {
				return chat.EmptyMessage, ErrTelegramLLMNotProvided
			},
		}
	}

	chatHistory := t.History
	if chatHistory == nil {
		chatHistory = &unimplementedChatHistoryReadWriter{}
	}

	userManager := t.User
	if userManager == nil {
		userManager = &unimplementedUserManager{}
	}

	errorLog := t.ErrorLog
	if errorLog == nil {
		t.ErrorLog = log.Default()
	}

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	bot.Debug = t.Debug

	unsupported := func(chatID int64) {
		if t.UnsupportedResponse != nil {
			if err := SendTelegramMessage(
				bot,
				chatID,
				t.UnsupportedResponse(),
			); err != nil {
				errorLog.Printf("failed to send unsupported message response: %v", err)
			}
		}
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case update := <-updates:
			if t.Handler == nil {
				continue
			}

			var (
				incoming chat.Message
				chatID   int64
				sender   *tgbotapi.User
				err      error
			)

			switch {
			case update.Message != nil:
				sender = update.Message.From
				chatID = update.Message.Chat.ID

				if update.Message.Text == "" {
					unsupported(update.Message.Chat.ID)
					continue
				}

				if update.Message.IsCommand() {
					incoming = chat.Message{
						Role: chat.RoleUser,
						Content: content.Command{
							Name: update.Message.Command(),
							Args: update.Message.CommandArguments(),
						},
					}
				} else {
					incoming = chat.Message{
						Role:    chat.RoleUser,
						Content: update.Message.Text,
					}
				}

			case update.CallbackQuery != nil &&
				update.CallbackQuery.Message != nil &&
				update.CallbackQuery.Message.Chat != nil:
				sender = update.CallbackQuery.From
				chatID = update.CallbackQuery.Message.Chat.ID

				var caption string
				if update.CallbackQuery.Message.ReplyMarkup != nil {
					for _, row := range update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard {
						if caption != "" {
							break
						}

						for _, col := range row {
							if col.CallbackData == nil {
								continue
							}

							if *col.CallbackData == update.CallbackQuery.Data {
								caption = col.Text
								break
							}
						}
					}
				}

				incoming = chat.Message{
					Role: chat.RoleUser,
					Content: content.SelectItem{
						Caption: caption,
						Data:    update.CallbackQuery.Data,
					},
				}

			default:
				errorLog.Printf("unsupported update: %v", update)
				continue
			}

			from, err := userManager.GetUser(ctx, sender.ID)
			if err != nil {
				if !errors.Is(err, user.ErrUserNotFound) {
					errorLog.Printf("failed to get user: %v", err)
					continue
				}

				from = chat.User{
					ID:        sender.ID,
					FirstName: sender.FirstName,
					LastName:  sender.LastName,
					UserName:  sender.UserName,
				}
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						errorLog.Printf("recovered from panic: %v", r)
					}
				}()

				t.Handler.Serve(
					ctx,
					&telegramResponseWriter{
						chatID:   chatID,
						sender:   bot,
						errorLog: errorLog,
					},
					&server.Request{
						ChatID:   chatID,
						Incoming: incoming,
						From:     from,

						Completer: chatCompletion,
						History:   chatHistory,
						User:      userManager,
						ErrorLog:  errorLog,
					})
			}()
		}
	}
}

// SendTelegramMessage отправляет сообщение через Telegram API.
func SendTelegramMessage(sender *tgbotapi.BotAPI, chatID int64, m chat.Message) (err error) {
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

	default:
		err = ErrTelegramUnsupportedMessageType
	}

	return
}

// telegramResponseWriter адаптирует отправку сообщений к интерфейсу ResponseWriter.
type telegramResponseWriter struct {
	chatID   int64
	sender   *tgbotapi.BotAPI
	errorLog *log.Logger
}

func (trw *telegramResponseWriter) WriteResponse(m chat.Message) error {
	if m.IsEmpty() {
		return nil
	}

	if err := SendTelegramMessage(trw.sender, trw.chatID, m); err != nil {
		trw.errorLog.Printf("failed to send message to chatID %d: %v", trw.chatID, err)
		return err
	}

	return nil
}

// unimplementedChatHistoryReadWriter предоставляет пустую реализацию интерфейса истории.
type unimplementedChatHistoryReadWriter struct{}

func (unimplementedChatHistoryReadWriter) ReadChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	return make([]chat.Message, 0), nil
}

func (unimplementedChatHistoryReadWriter) WriteChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	return nil
}

// unimplementedUserManager предоставляет пустую реализацию интерфейса менеджера пользователей.
type unimplementedUserManager struct{}

func (unimplementedUserManager) SaveUser(ctx context.Context, user chat.User) error {
	return nil
}

func (unimplementedUserManager) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	return chat.User{}, errors.New("user not found")
}
