package storage

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"go.etcd.io/bbolt"
)

// Имена бакетов для хранения в BoltDB.
var (
	chatBucket = []byte("chats")
	userBucket = []byte("users")
)

// Bolt реализует хранение в BoltDB (bbolt).
type Bolt struct {
	db   *bbolt.DB
	path string
}

// NewBolt создает новое хранилище BoltDB.
func NewBolt(path string) (*Bolt, error) {
	db, err := bbolt.Open(path, 0644, &bbolt.Options{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(chatBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(userBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Bolt{db: db, path: path}, nil
}

// Close закрывает соединение с базой данных.
func (b *Bolt) Close() error {
	return b.db.Close()
}

// GetChatHistory читает историю сообщений из BoltDB.
func (b *Bolt) GetChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	var msgs []chat.Message
	err := b.db.View(func(tx *bbolt.Tx) error {
		var (
			bucket = tx.Bucket(chatBucket)
			key    = []byte(strconv.FormatInt(chatID, 10))
			data   = bucket.Get(key)
		)
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &msgs)
	})

	if err != nil {
		return nil, err
	}

	if msgs == nil {
		return []chat.Message{}, nil
	}

	if limit == 0 || uint64(len(msgs)) <= limit {
		return msgs, nil
	}

	return msgs[uint64(len(msgs))-limit:], nil
}

// SaveChatHistory записывает историю сообщений в BoltDB.
func (b *Bolt) SaveChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	for _, msg := range msgs {
		if _, ok := msg.Content.(string); !ok {
			return ErrUnsupportedContentType
		}
	}

	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		var (
			bucket = tx.Bucket(chatBucket)
			key    = []byte(strconv.FormatInt(chatID, 10))
		)
		return bucket.Put(key, data)
	})
}

// GetUser возвращает пользователя из BoltDB по его ID.
func (b *Bolt) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	var user chat.User
	err := b.db.View(func(tx *bbolt.Tx) error {
		var (
			bucket = tx.Bucket(userBucket)
			key    = []byte(strconv.FormatInt(userID, 10))
			data   = bucket.Get(key)
		)
		if data == nil {
			return ErrUserNotFound
		}

		return json.Unmarshal(data, &user)
	})

	if err != nil {
		return chat.User{}, err
	}

	return user, nil
}

// SaveUser сохраняет пользователя в BoltDB.
func (b *Bolt) SaveUser(ctx context.Context, user chat.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		var (
			bucket = tx.Bucket(userBucket)
			key    = []byte(strconv.FormatInt(user.ID, 10))
		)
		return bucket.Put(key, data)
	})
}
