package history

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/muzykantov/health-gpt/chat"
)

// FileStorage реализует потокобезопасное хранение истории сообщений в файлах.
type FileStorage struct {
	mu  sync.RWMutex
	dir string
}

// NewFileStorage создает новое файловое хранилище истории сообщений.
func NewFileStorage(dir string) (*FileStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{dir: dir}, nil
}

// historyPath возвращает путь к файлу истории чата.
func (fs *FileStorage) historyPath(chatID int64) string {
	return filepath.Join(fs.dir, fmt.Sprintf("chat_%d.json", chatID))
}

// ReadChatHistory читает историю сообщений из файла.
func (fs *FileStorage) ReadChatHistory(
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

// WriteChatHistory записывает историю сообщений в файл.
func (fs *FileStorage) WriteChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(msgs, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.historyPath(chatID), data, 0644)
}
