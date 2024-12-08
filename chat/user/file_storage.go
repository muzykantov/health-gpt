package user

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/muzykantov/health-gpt/chat"
)

// FileStorage реализует потокобезопасное хранение пользователей в файлах.
type FileStorage struct {
	mu  sync.RWMutex
	dir string
}

// NewFileStorage создает новое файловое хранилище пользователей в указанной директории.
func NewFileStorage(dir string) (*FileStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{dir: dir}, nil
}

// userPath возвращает путь к файлу пользователя.
func (fs *FileStorage) userPath(userID int64) string {
	return filepath.Join(fs.dir, fmt.Sprintf("user_%d.json", userID))
}

// SaveUser сохраняет пользователя в файл.
func (fs *FileStorage) SaveUser(ctx context.Context, user chat.User) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.userPath(user.ID), data, 0644)
}

// GetUser возвращает пользователя из файла по его ID.
func (fs *FileStorage) GetUser(ctx context.Context, userID int64) (chat.User, error) {
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
