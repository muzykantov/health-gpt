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

// File реализует потокобезопасное хранение в файлах.
type File struct {
	mu  sync.RWMutex
	dir string
}

// NewFile создает новое файловое хранилище.
func NewFile(dir string) (*File, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &File{dir: dir}, nil
}

// GetChatHistory читает историю сообщений из файла.
func (f *File) GetChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, err := os.ReadFile(f.historyPath(chatID))
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
func (f *File) SaveChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, msg := range msgs {
		if _, ok := msg.Content.(string); !ok {
			return ErrUnsupportedContentType
		}
	}

	data, err := json.MarshalIndent(msgs, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.historyPath(chatID), data, 0644)
}

// GetUser возвращает пользователя из файла по его ID.
func (f *File) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, err := os.ReadFile(f.userPath(userID))
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
func (f *File) SaveUser(ctx context.Context, user chat.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.userPath(user.ID), data, 0644)
}

// historyPath возвращает путь к файлу истории чата.
func (f *File) historyPath(chatID int64) string {
	return filepath.Join(f.dir, fmt.Sprintf("chat_%d.json", chatID))
}

// userPath возвращает путь к файлу пользователя.
func (f *File) userPath(userID int64) string {
	return filepath.Join(f.dir, fmt.Sprintf("user_%d.json", userID))
}
