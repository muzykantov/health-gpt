package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/paulrzcz/go-gigachat"
)

var (
	ErrGigaChatUnsupportedRole        = errors.New("unsupported role")
	ErrGigaChatUnsupportedContentType = errors.New("unsupported content type")
	ErrGigaChatRequestFailed          = errors.New("request failed")
)

// Определение функции-опции для конфигурации клиента.
type GigaChatOption func(*GigaChat)

// Реализация интерфейса GigaChat.
type GigaChat struct {
	client *gigachat.Client
	// Конфигурационные параметры клиента
	model             string  // Модель для использования.
	temperature       float64 // Температура генерации (0.0-2.0).
	topP              float64 // Top-p сэмплирование (0.0-1.0).
	maxTokens         int64   // Максимальное количество токенов в ответе.
	repetitionPenalty float64 // Штраф за повторения.
}

// Установка модели для использования.
func GigaChatWithModel(model string) GigaChatOption {
	return func(c *GigaChat) {
		c.model = model
	}
}

// Установка температуры генерации.
func GigaChatWithTemperature(temperature float64) GigaChatOption {
	return func(c *GigaChat) {
		c.temperature = temperature
	}
}

// Установка параметра top-p сэмплирования.
func GigaChatWithTopP(topP float64) GigaChatOption {
	return func(c *GigaChat) {
		c.topP = topP
	}
}

// Установка максимального количества токенов.
func GigaChatWithMaxTokens(maxTokens int64) GigaChatOption {
	return func(c *GigaChat) {
		c.maxTokens = maxTokens
	}
}

// Установка штрафа за повторения.
func GigaChatWithRepetitionPenalty(penalty float64) GigaChatOption {
	return func(c *GigaChat) {
		c.repetitionPenalty = penalty
	}
}

// Создание нового экземпляра Client с заданными опциями.
func NewGigaChat(
	clientID,
	clientSecret string,
	opts ...GigaChatOption,
) (*GigaChat, error) {
	client, err := gigachat.NewInsecureClient(clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create GigaChat client: %w", err)
	}

	// Инициализация клиента со значениями по умолчанию.
	c := &GigaChat{
		client:            client,
		model:             "GigaChat-Max",
		temperature:       0.1,
		topP:              1.0,
		maxTokens:         1024,
		repetitionPenalty: 1.0,
	}

	// Применение переданных опций.
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Реализация интерфейса Completion.
func (gc *GigaChat) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	// Преобразование сообщений в формат GigaChat.
	gigaChatMessages := make([]gigachat.Message, len(msgs))
	for i, msg := range msgs {
		content, ok := msg.Content.(string)
		if !ok {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrGigaChatUnsupportedContentType,
				msg.Content,
			)
		}

		var role string
		switch msg.Role {
		case chat.RoleUser:
			role = "user"
		case chat.RoleAssistant:
			role = "assistant"
		case chat.RoleSystem:
			role = "system"
		default:
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %v",
				ErrGigaChatUnsupportedRole,
				msg.Role,
			)
		}

		gigaChatMessages[i] = gigachat.Message{
			Role:    role,
			Content: string(content),
		}
	}

	// Формирование запроса к GigaChat API.
	var (
		temp       = gc.temperature
		topP       = gc.topP
		maxTokens  = gc.maxTokens
		repPenalty = gc.repetitionPenalty
		req        = &gigachat.ChatRequest{
			Model:             gc.model,
			Messages:          gigaChatMessages,
			Temperature:       &temp,
			TopP:              &topP,
			MaxTokens:         &maxTokens,
			RepetitionPenalty: &repPenalty,
		}
	)

	// Выполнение авторизации.
	if err := gc.client.AuthWithContext(ctx); err != nil {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: gigachat auth failed: %w",
			ErrGigaChatRequestFailed,
			err,
		)
	}

	// Выполнение запроса к API.
	resp, err := gc.client.ChatWithContext(ctx, req)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: gigachat request failed: %w",
			ErrGigaChatRequestFailed,
			err,
		)
	}

	if len(resp.Choices) == 0 {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no choices in response",
			ErrGigaChatRequestFailed,
		)
	}

	// Преобразование ответа в формат.
	return chat.Message{
		Role:    chat.RoleAssistant,
		Content: resp.Choices[0].Message.Content,
	}, nil
}
