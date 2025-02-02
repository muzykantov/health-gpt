package server

import (
	"context"
	"log"

	"github.com/muzykantov/health-gpt/chat"
)

// ChatHistoryReadWriter объединяет чтение и запись истории диалога.
type ChatHistoryReadWriter interface {
	ReadChatHistory(ctx context.Context, chatID int64, limit uint64) ([]chat.Message, error)
	WriteChatHistory(ctx context.Context, chatID int64, msgs []chat.Message) error
}

// ChatCompleter генерирует ответы с помощью языковой модели.
type ChatCompleter interface {
	CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error)
}

// UserStorage хранит информацию о пользователях.
type UserStorage interface {
	SaveUser(ctx context.Context, user chat.User) error
	GetUser(ctx context.Context, userID int64) (chat.User, error)
}

// Request содержит входящее сообщение и сервисы для его обработки.
type Request struct {
	ChatID   int64        // Идентификатор чата.
	Incoming chat.Message // Входящее сообщение.
	From     chat.User    // Пользователь, отправивший входящее сообщение.

	Completer ChatCompleter         // Сервис генерации ответов.
	History   ChatHistoryReadWriter // Сервис чтения и записи истории диалога.
	User      UserStorage           // Сервис управления пользователями.

	ErrorLog *log.Logger // Сервис логирования ошибок
}

// ResponseWriter записывает ответное сообщение.
type ResponseWriter interface {
	WriteResponse(chat.Message) error
}

// Handler обрабатывает входящие сообщения и генерирует ответы.
type Handler interface {
	Serve(ctx context.Context, w ResponseWriter, r *Request)
}

// ----------------------------------------------------------------------

// HandlerFunc позволяет использовать функции как обработчики.
type HandlerFunc func(ctx context.Context, w ResponseWriter, r *Request)

// Serve реализует интерфейс Handler для HandlerFunc.
func (f HandlerFunc) Serve(ctx context.Context, w ResponseWriter, r *Request) {
	f(ctx, w, r)
}
