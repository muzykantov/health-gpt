-- +goose Up
-- +goose StatementBegin

CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    sender_role INTEGER NOT NULL,
    content_type VARCHAR(50) NOT NULL, 
    content JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_messages_chat_created 
    ON chat_messages (chat_id, created_at DESC);

CREATE INDEX idx_chat_messages_content_type 
    ON chat_messages (content_type);

CREATE INDEX idx_chat_messages_content 
    ON chat_messages USING GIN (content);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS chat_messages;

-- +goose StatementEnd