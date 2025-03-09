package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/muzykantov/health-gpt/chat"
)

// FS реализует потокобезопасное хранение в файлах.
type FS struct {
	mu  sync.RWMutex
	dir string
}

// NewFS создает новое файловое хранилище.
func NewFS(dir string) (*FS, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FS{dir: dir}, nil
}

// GetChatHistory читает историю сообщений из файла.
func (fs *FS) GetChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.historyPath(chatID))
	if os.IsNotExist(err) {
		return []chat.Message{}, nil
	}
	if err != nil {
		return nil, err
	}

	var msgs []chat.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, err
	}

	if uint64(len(msgs)) <= limit || limit == 0 {
		return msgs, nil
	}
	return msgs[uint64(len(msgs))-limit:], nil
}

// SaveChatHistory записывает историю сообщений в файл.
func (fs *FS) SaveChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, msg := range msgs {
		if _, ok := msg.Content.(string); !ok {
			return ErrUnsupportedContentType
		}
	}

	data, err := json.MarshalIndent(msgs, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.historyPath(chatID), data, 0644)
}

// GetUser возвращает пользователя из файла по его ID.
func (fs *FS) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.userPath(userID))
	if os.IsNotExist(err) {
		return chat.User{}, ErrUserNotFound
	}
	if err != nil {
		return chat.User{}, err
	}

	var user chat.User
	if err := json.Unmarshal(data, &user); err != nil {
		return chat.User{}, err
	}

	return user, nil
}

// SaveUser сохраняет пользователя в файл.
func (fs *FS) SaveUser(ctx context.Context, user chat.User) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.userPath(user.ID), data, 0644)
}

// historyPath возвращает путь к файлу истории чата.
func (fs *FS) historyPath(chatID int64) string {
	return filepath.Join(fs.dir, fmt.Sprintf("chat_%d.json", chatID))
}

// userPath возвращает путь к файлу пользователя.
func (fs *FS) userPath(userID int64) string {
	return filepath.Join(fs.dir, fmt.Sprintf("user_%d.json", userID))
}
