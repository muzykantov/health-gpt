package chat

import "fmt"

// Message объединяет роль и содержимое сообщения.
type Message struct {
	Role
	Content any
}

// IsEmpty проверяет, является ли сообщение пустым.
func (m Message) IsEmpty() bool {
	return m.Role == RoleUndefined && m.Content == nil
}

// Equals проверяет, равны ли два сообщения.
func (m Message) Equals(other Message) bool {
	return m.Role == other.Role && m.Content == other.Content
}

// String возвращает строковое представление сообщения.
func (m Message) String() string {
	return fmt.Sprintf("%s: %v", m.Role, m.Content)
}

// NewMessage создает новое сообщение.
func NewMessage(role Role, content any) Message {
	return Message{
		Role:    role,
		Content: content,
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
	Role:    RoleUndefined,
	Content: nil,
}
