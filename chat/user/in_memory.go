package user

import (
	"context"
	"sync"

	"github.com/muzykantov/health-gpt/chat"
)

// InMemory реализует потокобезопасное хранение пользователей в памяти.
type InMemory struct {
	mu    sync.RWMutex
	users map[int64]chat.User
}

// NewInMemory создает новое хранилище пользователей.
func NewInMemory() *InMemory {
	return &InMemory{
		users: make(map[int64]chat.User),
	}
}

// SaveUser сохраняет пользователя в хранилище.
func (im *InMemory) SaveUser(ctx context.Context, user chat.User) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.users[user.ID] = user
	return nil
}

// GetUser возвращает пользователя по его ID.
func (im *InMemory) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	user, exists := im.users[userID]
	if !exists {
		return chat.User{}, ErrUserNotFound
	}

	return user, nil
}
