package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/chat/storage"
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
	Storage             server.DataStorage
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

	dataStorage := t.Storage
	if dataStorage == nil {
		dataStorage = &unimplementedDataStorage{}
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
			if err := SendMessage(
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
						Sender: chat.RoleUser,
						Content: content.Command{
							Name: update.Message.Command(),
							Args: update.Message.CommandArguments(),
						},
						CreatedAt: chat.Now().UTC(),
					}
				} else {
					incoming = chat.Message{
						Sender:    chat.RoleUser,
						Content:   update.Message.Text,
						CreatedAt: chat.Now().UTC(),
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
					Sender: chat.RoleUser,
					Content: content.SelectItem{
						Caption: caption,
						Data:    update.CallbackQuery.Data,
					},
					CreatedAt: chat.Now().UTC(),
				}

			default:
				errorLog.Printf("unsupported update: %v", update)
				continue
			}

			from, err := dataStorage.GetUser(ctx, sender.ID)
			if err != nil {
				if !errors.Is(err, storage.ErrUserNotFound) {
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
						Storage:   dataStorage,
						ErrorLog:  errorLog,
					})
			}()
		}
	}
}

// telegramResponseWriter адаптирует отправку сообщений к интерфейсу ResponseWriter.
type telegramResponseWriter struct {
	chatID   int64
	sender   *tgbotapi.BotAPI
	errorLog *log.Logger
}

func (w *telegramResponseWriter) WriteResponse(m chat.Message) error {
	if m.IsEmpty() {
		return nil
	}

	if err := SendMessage(w.sender, w.chatID, m); err != nil {
		w.errorLog.Printf("failed to send message to chatID %d: %v", w.chatID, err)
		return err
	}

	return nil
}

// unimplementedDataStorage предоставляет пустую реализацию интерфейса истории.
type unimplementedDataStorage struct{}

func (unimplementedDataStorage) GetChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	return make([]chat.Message, 0), nil
}

func (unimplementedDataStorage) SaveChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	return nil
}

func (unimplementedDataStorage) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	return chat.User{}, errors.New("user not found")
}

func (unimplementedDataStorage) SaveUser(ctx context.Context, user chat.User) error {
	return nil
}
