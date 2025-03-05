package server

import (
	"context"
	"log"

	"github.com/muzykantov/health-gpt/chat"
)

// ChatCompleter генерирует ответы с помощью языковой модели.
type ChatCompleter interface {
	CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error)
}

// ChatHistoryStorage объединяет чтение и запись истории диалога.
type ChatHistoryStorage interface {
	GetChatHistory(ctx context.Context, chatID int64, limit uint64) ([]chat.Message, error)
	SaveChatHistory(ctx context.Context, chatID int64, msgs []chat.Message) error
}

// UserStorage хранит информацию о пользователях.
type UserStorage interface {
	GetUser(ctx context.Context, userID int64) (chat.User, error)
	SaveUser(ctx context.Context, user chat.User) error
}

// Storage отвечает за получение и хранение данных.
type DataStorage interface {
	ChatHistoryStorage
	UserStorage
}

// Request содержит входящее сообщение и сервисы для его обработки.
type Request struct {
	ChatID   int64        // Идентификатор чата.
	Incoming chat.Message // Входящее сообщение.
	From     chat.User    // Пользователь, отправивший входящее сообщение.

	Completer ChatCompleter // Сервис генерации ответов.
	Storage   DataStorage

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
