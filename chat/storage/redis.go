package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/redis/go-redis/v9"
)

// Ключи Redis
const (
	prefixChatHistory = "chat:%d:history"
	prefixUser        = "user:%d"
)

// Redis реализует хранение в Redis.
type Redis struct {
	client *redis.Client
	// Время жизни записей в Redis (0 - бессрочно)
	expiration time.Duration
}

// NewRedis создает новое Redis хранилище.
func NewRedis(options *redis.Options, expiration time.Duration) (*Redis, error) {
	client := redis.NewClient(options)

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{
		client:     client,
		expiration: expiration,
	}, nil
}

// GetChatHistory читает историю сообщений из Redis.
func (r *Redis) GetChatHistory(
	ctx context.Context,
	chatID int64,
	limit uint64,
) ([]chat.Message, error) {
	key := fmt.Sprintf(prefixChatHistory, chatID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return []chat.Message{}, nil
		}
		return nil, err
	}

	var msgs []chat.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, err
	}

	// Если limit=0 или количество сообщений меньше лимита,
	// то возвращаем все сообщения
	if limit == 0 || uint64(len(msgs)) <= limit {
		return msgs, nil
	}

	// Иначе возвращаем только последние limit сообщений
	return msgs[uint64(len(msgs))-limit:], nil
}

// SaveChatHistory записывает историю сообщений в Redis.
func (r *Redis) SaveChatHistory(
	ctx context.Context,
	chatID int64,
	msgs []chat.Message,
) error {
	// Проверяем, что все сообщения имеют контент типа string
	for _, msg := range msgs {
		if _, ok := msg.Content.(string); !ok {
			return ErrUnsupportedContentType
		}
	}

	key := fmt.Sprintf(prefixChatHistory, chatID)

	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.expiration).Err()
}

// GetUser возвращает пользователя из Redis по его ID.
func (r *Redis) GetUser(ctx context.Context, userID int64) (chat.User, error) {
	key := fmt.Sprintf(prefixUser, userID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return chat.User{}, ErrUserNotFound
		}
		return chat.User{}, err
	}

	var user chat.User
	if err := json.Unmarshal(data, &user); err != nil {
		return chat.User{}, err
	}

	return user, nil
}

// SaveUser сохраняет пользователя в Redis.
func (r *Redis) SaveUser(ctx context.Context, user chat.User) error {
	key := fmt.Sprintf(prefixUser, user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.expiration).Err()
}

// Close закрывает соединение с Redis.
func (r *Redis) Close() error {
	return r.client.Close()
}
