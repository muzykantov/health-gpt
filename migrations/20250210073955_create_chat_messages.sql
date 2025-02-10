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

-- Добавляем комментарии к таблице и колонкам для документации схемы
COMMENT ON TABLE chat_messages IS 'Хранит историю сообщений чата';
COMMENT ON COLUMN chat_messages.chat_id IS 'Идентификатор чата';
COMMENT ON COLUMN chat_messages.sender_role IS 'Роль отправителя: 0 - не определено, 1 - пользователь, 2 - ассистент, 3 - система';
COMMENT ON COLUMN chat_messages.content_type IS 'Тип контента: text, command, commands, select, menu_item';
COMMENT ON COLUMN chat_messages.content IS 'Содержимое сообщения в формате JSON';
COMMENT ON COLUMN chat_messages.created_at IS 'Время создания сообщения';

-- Создаем индексы для оптимизации запросов
-- Основной индекс для поиска сообщений чата с сортировкой по времени
CREATE INDEX idx_chat_messages_chat_created 
    ON chat_messages (chat_id, created_at DESC);

-- Индекс для аналитики по типам контента
CREATE INDEX idx_chat_messages_content_type 
    ON chat_messages (content_type);

-- Индекс для поиска по содержимому JSON
CREATE INDEX idx_chat_messages_content 
    ON chat_messages USING GIN (content);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS chat_messages;

-- +goose StatementEnd