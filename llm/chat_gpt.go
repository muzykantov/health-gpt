package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/proxy"
)

var (
	ErrChatGPTUnsupportedRole        = errors.New("unsupported role")
	ErrChatGPTUnsupportedContentType = errors.New("unsupported content type")
	ErrChatGPTRequestFailed          = errors.New("request failed")
)

// ChatGPTOption определяет функцию-опцию для конфигурации клиента.
type ChatGPTOption func(*ChatGPT)

// ChatGPT реализует клиент для взаимодействия с ChatGPT.
type ChatGPT struct {
	client *openai.Client
	// Конфигурационные параметры.
	model       string  // Модель для использования.
	temperature float32 // Температура генерации (0.0-2.0).
	topP        float32 // Top-p сэмплирование (0.0-1.0).
	maxTokens   int     // Максимальное количество токенов в ответе.
	socksProxy  string  // Адрес прокси-сокса.
}

// Установка модели для использования.
func ChatGPTWithModel(model string) ChatGPTOption {
	return func(c *ChatGPT) {
		c.model = model
	}
}

// Установка температуры генерации.
func ChatGPTWithTemperature(temperature float32) ChatGPTOption {
	return func(c *ChatGPT) {
		c.temperature = temperature
	}
}

// Установка параметра top-p сэмплирования.
func ChatGPTWithTopP(topP float32) ChatGPTOption {
	return func(c *ChatGPT) {
		c.topP = topP
	}
}

// Установка максимального количества токенов.
func ChatGPTWithMaxTokens(maxTokens int) ChatGPTOption {
	return func(c *ChatGPT) {
		c.maxTokens = maxTokens
	}
}

func ChatGPTWithSocksProxy(socksProxy string) ChatGPTOption {
	return func(c *ChatGPT) {
		c.socksProxy = socksProxy
	}
}

// NewChatGPT создает новый экземпляр ChatGPT с заданными опциями.
func NewChatGPT(apiKey string, opts ...ChatGPTOption) (*ChatGPT, error) {
	// Инициализация клиента со значениями по умолчанию.
	c := &ChatGPT{
		client:      openai.NewClient(apiKey),
		model:       openai.GPT4o,
		temperature: 0.1,
		topP:        1.0,
		maxTokens:   1024,
	}

	// Применение переданных опций.
	for _, opt := range opts {
		opt(c)
	}

	cfg := openai.DefaultConfig(apiKey)
	if c.socksProxy != "" {
		dialer, err := proxy.SOCKS5("tcp", c.socksProxy, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}

		transport := &http.Transport{
			DialContext:       dialContext,
			DisableKeepAlives: true,
		}

		cfg.HTTPClient = &http.Client{Transport: transport}
	}

	c.client = openai.NewClientWithConfig(cfg)

	return c, nil
}

// CompleteChat реализует интерфейс Completion.
func (c *ChatGPT) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	// Преобразование сообщений в формат OpenAI.
	openAIMessages := make([]openai.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		content, ok := msg.Content.(string)
		if !ok {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrChatGPTUnsupportedContentType,
				msg.Content,
			)
		}

		var role string
		switch msg.Sender {
		case chat.RoleUser:
			role = openai.ChatMessageRoleUser
		case chat.RoleAssistant:
			role = openai.ChatMessageRoleAssistant
		case chat.RoleSystem:
			role = openai.ChatMessageRoleSystem
		default:
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %v",
				ErrChatGPTUnsupportedRole,
				msg.Sender,
			)
		}

		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		}
	}

	// Формирование запроса к ChatGPT API.
	req := openai.ChatCompletionRequest{
		Model:               c.model,
		Messages:            openAIMessages,
		Temperature:         c.temperature,
		TopP:                c.topP,
		MaxCompletionTokens: c.maxTokens,
	}

	// Выполнение запроса к API.
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: chatgpt request failed: %w",
			ErrChatGPTRequestFailed,
			err,
		)
	}

	if len(resp.Choices) == 0 {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no choices in response",
			ErrChatGPTRequestFailed,
		)
	}

	// Преобразование ответа в нужный формат.
	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: resp.Choices[0].Message.Content,
	}, nil
}
