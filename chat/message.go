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

// EmptyMessage представляет пустое сообщение.
var EmptyMessage = Message{
	Role:    RoleUndefined,
	Content: nil,
}
