package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
)

type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeCommand  ContentType = "command"
	ContentTypeCommands ContentType = "commands"
	ContentTypeSelect   ContentType = "select"
	ContentTypeMenuItem ContentType = "menu_item"
)

var ErrUnsupportedContentType = errors.New("content type not supported")

type DBStorage struct {
	db *sql.DB
}

// NewDB создает новое хранилище истории чата.
func NewDB(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

// contentType определяет тип контента сообщения.
// Возвращает ошибку если тип не поддерживается.
func contentType(v any) (ContentType, error) {
	switch v.(type) {
	case string:
		return ContentTypeText, nil
	case content.Command:
		return ContentTypeCommand, nil
	case content.Commands:
		return ContentTypeCommands, nil
	case content.Select:
		return ContentTypeSelect, nil
	case content.SelectItem:
		return ContentTypeMenuItem, nil
	default:
		return "", ErrUnsupportedContentType
	}
}

// ReadHistory возвращает последние сообщения из истории чата.
// Сообщения возвращаются в хронологическом порядке.
func (s *DBStorage) ReadHistory(ctx context.Context, chatID int64, limit uint64) ([]chat.Message, error) {
	query := `
		SELECT role, content_type, content 
    	FROM chat_messages 
        WHERE chat_id = $1 
        ORDER BY created_at DESC
	`

	if limit > 0 {
		query += " LIMIT $2"
	}

	var (
		rows *sql.Rows
		err  error
	)

	if limit > 0 {
		rows, err = s.db.QueryContext(ctx, query, chatID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, query, chatID)
	}

	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var messages []chat.Message
	for rows.Next() {
		var (
			msg         chat.Message
			roleInt     int
			contentType string
			contentJSON []byte
		)

		if err := rows.Scan(&roleInt, &contentType, &contentJSON); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		msg.Role = chat.Role(roleInt)

		switch ContentType(contentType) {
		case ContentTypeText:
			var text string
			if err := json.Unmarshal(contentJSON, &text); err != nil {
				return nil, fmt.Errorf("unmarshal text content: %w", err)
			}
			msg.Content = text

		case ContentTypeCommand:
			var cmd content.Command
			if err := json.Unmarshal(contentJSON, &cmd); err != nil {
				return nil, fmt.Errorf("unmarshal command content: %w", err)
			}
			msg.Content = cmd

		case ContentTypeCommands:
			var cmds content.Commands
			if err := json.Unmarshal(contentJSON, &cmds); err != nil {
				return nil, fmt.Errorf("unmarshal commands content: %w", err)
			}
			msg.Content = cmds

		case ContentTypeSelect:
			var sel content.Select
			if err := json.Unmarshal(contentJSON, &sel); err != nil {
				return nil, fmt.Errorf("unmarshal select content: %w", err)
			}
			msg.Content = sel

		case ContentTypeMenuItem:
			var item content.SelectItem
			if err := json.Unmarshal(contentJSON, &item); err != nil {
				return nil, fmt.Errorf("unmarshal menu item content: %w", err)
			}
			msg.Content = item

		default:
			return nil, fmt.Errorf("unknown content type: %s", contentType)
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	slices.Reverse(messages)

	return messages, nil
}

// WriteHistory сохраняет сообщения в историю чата.
func (s *DBStorage) WriteHistory(ctx context.Context, chatID int64, msgs []chat.Message) error {
	query := `
		INSERT INTO chat_messages (chat_id, role, content_type, content)
		VALUES ($1, $2, $3, $4::jsonb)
	`

	if len(msgs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, msg := range msgs {
		contentType, err := contentType(msg.Content)
		if err != nil {
			return fmt.Errorf("message %d: %w", i, err)
		}

		contentJSON, err := json.Marshal(msg.Content)
		if err != nil {
			return fmt.Errorf("message %d: marshal content: %w", i, err)
		}

		if _, err := stmt.ExecContext(ctx,
			chatID,
			int(msg.Role),
			string(contentType),
			contentJSON,
		); err != nil {
			return fmt.Errorf("message %d: insert: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
