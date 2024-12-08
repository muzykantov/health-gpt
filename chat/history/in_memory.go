package history

import (
	"context"
	"sync"

	"github.com/muzykantov/health-gpt/chat"
)

// InMemory реализует потокобезопасное хранение истории сообщений в памяти.
type InMemory struct {
	mu      sync.RWMutex
	history map[int64][]chat.Message
}

// NewInMemory создает новое хранилище сообщений.
func NewInMemory() *InMemory {
	return &InMemory{
		history: make(map[int64][]chat.Message),
	}
}

// ReadChatHistory читает историю сообщений из хранилища.
func (m *InMemory) ReadChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs, ok := m.history[chatID]
	if !ok {
		return []chat.Message{}, nil
	}

	if uint64(len(msgs)) <= limit || limit == 0 {
		return msgs, nil
	}
	return msgs[uint64(len(msgs))-limit:], nil
}

// WriteChatHistory записывает историю сообщений в хранилище.
func (m *InMemory) WriteChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history[chatID] = msgs
	return nil
}
