package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
)

// ContentType represents the type of message content.
// Different content types determine how the message content
// will be serialized and deserialized.
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeCommand  ContentType = "command"
	ContentTypeCommands ContentType = "commands"
	ContentTypeSelect   ContentType = "select"
	ContentTypeMenuItem ContentType = "menu_item"
)

// ErrUnsupportedContentType indicates that message content has type
// which cannot be stored in the database.
var ErrUnsupportedContentType = errors.New("content type not supported")

// DBStorage implements chat history storage using SQL database.
// It handles serialization and deserialization of different message types
// and maintains message ordering.
type DBStorage struct {
	db *sql.DB
}

// NewDB creates new chat history storage instance.
func NewDB(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

// contentType determines the type of message content.
// Returns ErrUnsupportedContentType if content type cannot be stored.
func contentType(v any) (ContentType, error) {
	switch v := v.(type) {
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
		return "", fmt.Errorf("%w: type %T", ErrUnsupportedContentType, v)
	}
}

// ReadHistory returns last messages from chat history in chronological order.
// If limit is 0, returns all messages.
func (s *DBStorage) ReadHistory(ctx context.Context, chatID int64, limit uint64) ([]chat.Message, error) {
	query := `
        SELECT sender_role, content_type, content, created_at 
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
			createdAt   time.Time
		)

		if err := rows.Scan(&roleInt, &contentType, &contentJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		msg.Sender = chat.Role(roleInt)
		msg.CreatedAt = createdAt

		// Parse content based on its type
		switch ContentType(contentType) {
		case ContentTypeText:
			var text string
			if err := json.Unmarshal(contentJSON, &text); err != nil {
				return nil, fmt.Errorf("unmarshal text content: %w (content: %s)", err, string(contentJSON))
			}
			msg.Content = text

		case ContentTypeCommand:
			var cmd content.Command
			if err := json.Unmarshal(contentJSON, &cmd); err != nil {
				return nil, fmt.Errorf("unmarshal command content: %w (content: %s)", err, string(contentJSON))
			}
			msg.Content = cmd

		case ContentTypeCommands:
			var cmds content.Commands
			if err := json.Unmarshal(contentJSON, &cmds); err != nil {
				return nil, fmt.Errorf("unmarshal commands content: %w (content: %s)", err, string(contentJSON))
			}
			msg.Content = cmds

		case ContentTypeSelect:
			var sel content.Select
			if err := json.Unmarshal(contentJSON, &sel); err != nil {
				return nil, fmt.Errorf("unmarshal select content: %w (content: %s)", err, string(contentJSON))
			}
			msg.Content = sel

		case ContentTypeMenuItem:
			var item content.SelectItem
			if err := json.Unmarshal(contentJSON, &item); err != nil {
				return nil, fmt.Errorf("unmarshal menu item content: %w (content: %s)", err, string(contentJSON))
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

	// Reverse messages to get chronological order
	slices.Reverse(messages)

	return messages, nil
}

// WriteHistory saves messages to chat history.
// All messages must have valid content type and non-nil content.
// Messages are saved in a single transaction - either all messages
// are saved or none.
func (s *DBStorage) WriteHistory(ctx context.Context, chatID int64, msgs []chat.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	query := `
        INSERT INTO chat_messages (chat_id, sender_role, content_type, content, created_at)
        VALUES ($1, $2, $3, $4::jsonb, $5)
    `

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("rollback transaction: %v", err)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, msg := range msgs {
		if msg.Content == nil {
			return fmt.Errorf("message %d: content is nil", i)
		}

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
			int(msg.Sender),
			string(contentType),
			contentJSON,
			msg.CreatedAt,
		); err != nil {
			return fmt.Errorf("message %d: insert: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
