package chat

// Role определяет отправителя сообщения.
type Role int

const (
	RoleUndefined Role = iota // Не определено.
	RoleUser                  // Пользователь.
	RoleAssistant             // Ассистент.
	RoleSystem                // Системное сообщение.
)

// String возвращает строковое представление роли.
func (r Role) String() string {
	switch r {
	case RoleUndefined:
		return "undefined"
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	case RoleSystem:
		return "system"
	default:
		return "unknown"
	}
}
