package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/chat/storage"
	"github.com/muzykantov/health-gpt/llm"
	"github.com/muzykantov/health-gpt/metrics"
	"github.com/muzykantov/health-gpt/server"
)

// Main Telegram API errors
var (
	ErrTelegramTokenNotProvided       = errors.New("telegram token not provided")
	ErrTelegramLLMNotProvided         = errors.New("telegram llm not provided")
	ErrTelegramUnsupportedMessageType = errors.New("telegram unsupported message type")
	ErrTelegramInvalidMessageContent  = errors.New("telegram invalid message content")
)

// Server manages interaction with Telegram Bot API
type Server struct {
	Token               string
	Handler             server.Handler
	Completion          server.ChatCompleter
	Storage             server.DataStorage
	Debug               bool
	UnsupportedResponse func() chat.Message
	Log                 *log.Logger

	// For tracking active users
	activeUsers   map[int64]bool
	activeUsersMu sync.Mutex
}

// ListenAndServe starts the main message processing loop
func (t *Server) ListenAndServe(ctx context.Context) error {
	if t.Token == "" {
		return ErrTelegramTokenNotProvided
	}

	// Initialize active users tracker
	t.activeUsers = make(map[int64]bool)

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

	cache := expirable.NewLRU[string, any](0, nil, time.Hour)

	dataStorage := t.Storage
	if dataStorage == nil {
		dataStorage = &unimplementedDataStorage{}
	}

	logger := t.Log
	if logger == nil {
		t.Log = log.Default()
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
				logger.Printf("failed to send unsupported message response: %v", err)
				metrics.RecordTelegramError("unsupported_response")
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
				incoming    chat.Message
				chatID      int64
				sender      *tgbotapi.User
				err         error
				messageType string
			)

			switch {
			case update.Message != nil:
				sender = update.Message.From
				chatID = update.Message.Chat.ID

				if update.Message.Text == "" {
					unsupported(update.Message.Chat.ID)
					metrics.RecordTelegramMessage("unsupported")
					continue
				}

				if update.Message.IsCommand() {
					messageType = "command"
					metrics.RecordTelegramMessage("command")
					incoming = chat.MsgU(
						content.Command{
							Name: update.Message.Command(),
							Args: update.Message.CommandArguments(),
						},
					)
				} else {
					messageType = "text"
					metrics.RecordTelegramMessage("text")
					incoming = chat.MsgU(update.Message.Text)
				}

			case update.CallbackQuery != nil &&
				update.CallbackQuery.Message != nil &&
				update.CallbackQuery.Message.Chat != nil:
				sender = update.CallbackQuery.From
				chatID = update.CallbackQuery.Message.Chat.ID
				messageType = "callback"
				metrics.RecordTelegramMessage("callback")

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

				incoming = chat.MsgU(content.SelectItem{
					Caption: caption,
					Data:    update.CallbackQuery.Data,
				})

			default:
				logger.Printf("unsupported update: %v", update)
				metrics.RecordTelegramMessage("unknown")
				continue
			}

			// Track unique users
			isNewUser := false
			t.activeUsersMu.Lock()
			if !t.activeUsers[sender.ID] {
				t.activeUsers[sender.ID] = true
				isNewUser = true
				// Update active users counter
				metrics.UpdateActiveUsers(len(t.activeUsers))
			}
			t.activeUsersMu.Unlock()

			// Record user session
			metrics.RecordUserSession(isNewUser)

			from, err := dataStorage.GetUser(ctx, sender.ID)
			if err != nil {
				if !errors.Is(err, storage.ErrUserNotFound) {
					logger.Printf("failed to get user: %v", err)
					metrics.RecordTelegramError("get_user")
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
						logger.Printf("recovered from panic: %v", r)
						metrics.RecordTelegramError("panic")
					}
				}()

				start := time.Now()

				t.Handler.Serve(
					ctx,
					&telegramResponseWriter{
						chatID:      chatID,
						sender:      bot,
						log:         logger,
						messageType: messageType,
						startTime:   start,
					},
					&server.Request{
						ChatID:   chatID,
						Incoming: incoming,
						From:     from,

						Completer: chatCompletion,
						Storage:   dataStorage,
						Cache:     cache,
						Log:       logger,
					})
			}()
		}
	}
}

// telegramResponseWriter adapts message sending to the ResponseWriter interface
type telegramResponseWriter struct {
	chatID      int64
	sender      *tgbotapi.BotAPI
	log         *log.Logger
	messageType string
	startTime   time.Time
}

func (w *telegramResponseWriter) WriteResponse(m chat.Message) error {
	if m.IsEmpty() {
		return nil
	}

	if err := SendMessage(w.sender, w.chatID, m); err != nil {
		w.log.Printf("failed to send message to chatID %d: %v", w.chatID, err)
		metrics.RecordTelegramError("send_message")
		return err
	}

	// Record response time
	responseTime := time.Since(w.startTime).Seconds()
	metrics.ObserveTelegramResponseTime(w.messageType, responseTime)

	return nil
}

// unimplementedDataStorage provides an empty implementation of the history interface
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
