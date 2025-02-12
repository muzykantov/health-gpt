package chat

import (
	"fmt"
	"time"
)

// Now можно переопределить в тестах.
var Now = time.Now

// Message объединяет роль и содержимое сообщения.
type Message struct {
	Sender    Role
	Content   any
	CreatedAt time.Time
}

// IsEmpty проверяет, является ли сообщение пустым.
func (m Message) IsEmpty() bool {
	return m.Sender == RoleUndefined && m.Content == nil
}

// Equals проверяет, равны ли два сообщения.
func (m Message) Equals(other Message) bool {
	return m.Sender == other.Sender &&
		m.Content == other.Content &&
		m.CreatedAt.Equal(other.CreatedAt)
}

// String возвращает строковое представление сообщения.
func (m Message) String() string {
	return fmt.Sprintf("%s: %v", m.Sender, m.Content)
}

// NewMessage создает новое сообщение.
func NewMessage(role Role, content any) Message {
	return Message{
		Sender:    role,
		Content:   content,
		CreatedAt: Now().UTC(),
	}
}

// MsgA создает новое сообщение от ассистента.
func MsgA(content any) Message {
	return NewMessage(RoleAssistant, content)
}

// MsgS создает новое системное сообщение.
func MsgS(content any) Message {
	return NewMessage(RoleSystem, content)
}

// MsgU создает новое сообщение от пользователя.
func MsgU(content any) Message {
	return NewMessage(RoleUser, content)
}

// MsgAf создает новое сообщение от ассистента с форматированием.
func MsgAf(format string, a ...any) Message {
	return MsgA(fmt.Sprintf(format, a...))
}

// MsgSf создает новое системное сообщение с форматированием.
func MsgSf(format string, a ...any) Message {
	return MsgS(fmt.Sprintf(format, a...))
}

// MsgUf создает новое сообщение от пользователя с форматированием.
func MsgUf(format string, a ...any) Message {
	return MsgU(fmt.Sprintf(format, a...))
}

// EmptyMessage представляет пустое сообщение.
var EmptyMessage = Message{
	Sender:  RoleUndefined,
	Content: nil,
}
